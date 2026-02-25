package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
)

type GraphQLClient struct {
	url    string
	token  string
	client *http.Client
}

func NewGraphQLClient(baseURL, token string) *GraphQLClient {
	return &GraphQLClient{
		url:    strings.TrimRight(baseURL, "/") + "/api/graphql",
		token:  token,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type gqlPipelineNode struct {
	IID       string `json:"iid"`
	ID        string `json:"id"`
	Ref       string `json:"ref"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	Duration  *int   `json:"duration"`
}

type gqlPipelineEdges struct {
	Nodes []gqlPipelineNode `json:"nodes"`
}

type gqlProjectResult struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	FullPath      string           `json:"fullPath"`
	WebURL        string           `json:"webUrl"`
	Pipelines     gqlPipelineEdges `json:"pipelines"`
}

func (c *GraphQLClient) do(ctx context.Context, req gqlRequest) (*gqlResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphql: HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp gqlResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("graphql: unmarshal: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		msgs := make([]string, len(gqlResp.Errors))
		for i, e := range gqlResp.Errors {
			msgs[i] = e.Message
		}
		return nil, fmt.Errorf("graphql: %s", strings.Join(msgs, "; "))
	}

	return &gqlResp, nil
}

// LoadAllPipelines fetches pipelines for all projects in a single GraphQL query.
func (c *GraphQLClient) LoadAllPipelines(ctx context.Context, projectPaths []string, perProject int) ([]entity.Pipeline, error) {
	if len(projectPaths) == 0 {
		return nil, nil
	}

	// Build query with aliases: p0, p1, p2...
	var fragments []string
	for i, path := range projectPaths {
		fragments = append(fragments, fmt.Sprintf(
			`p%d: project(fullPath: %q) {
				id name fullPath webUrl
				pipelines(first: %d, sort: CREATED_DESC) {
					nodes { iid id ref status createdAt duration }
				}
			}`, i, path, perProject))
	}
	query := "query { " + strings.Join(fragments, "\n") + " }"

	resp, err := c.do(ctx, gqlRequest{Query: query})
	if err != nil {
		return nil, err
	}

	// Parse dynamic aliases
	var data map[string]json.RawMessage
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}

	var all []entity.Pipeline
	for i, path := range projectPaths {
		key := fmt.Sprintf("p%d", i)
		raw, ok := data[key]
		if !ok || string(raw) == "null" {
			continue
		}

		var proj gqlProjectResult
		if err := json.Unmarshal(raw, &proj); err != nil {
			return nil, fmt.Errorf("parsing project %s: %w", path, err)
		}

		projectID := extractNumericID(proj.ID)

		for _, node := range proj.Pipelines.Nodes {
			createdAt, _ := time.Parse(time.RFC3339, node.CreatedAt)
			dur := 0
			if node.Duration != nil {
				dur = *node.Duration
			}

			all = append(all, entity.Pipeline{
				ID:          extractNumericID(node.ID),
				ProjectID:   projectID,
				ProjectPath: proj.FullPath,
				Ref:         node.Ref,
				Status:      mapGQLStatus(node.Status),
				CreatedAt:   createdAt,
				Duration:    dur,
			})
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	return all, nil
}

// mapGQLStatus converts GraphQL pipeline status (UPPERCASE) to our domain status (lowercase).
func mapGQLStatus(s string) valueobject.PipelineStatus {
	switch strings.ToUpper(s) {
	case "RUNNING":
		return valueobject.PipelineRunning
	case "PENDING":
		return valueobject.PipelinePending
	case "SUCCESS":
		return valueobject.PipelineSuccess
	case "FAILED":
		return valueobject.PipelineFailed
	case "CANCELED":
		return valueobject.PipelineCanceled
	case "SKIPPED":
		return valueobject.PipelineSkipped
	case "MANUAL":
		return valueobject.PipelineManual
	case "CREATED":
		return valueobject.PipelineCreated
	default:
		return valueobject.PipelineStatus(strings.ToLower(s))
	}
}

// extractNumericID extracts the numeric part from a GitLab GraphQL global ID.
// e.g. "gid://gitlab/Pipeline/12345" -> 12345
func extractNumericID(gid string) int {
	parts := strings.Split(gid, "/")
	if len(parts) == 0 {
		return 0
	}
	last := parts[len(parts)-1]
	var id int
	fmt.Sscanf(last, "%d", &id)
	return id
}
