package mcp

import (
	"context"
	"fmt"
	"io"

	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool input types

type ListProjectsInput struct{}

type ListPipelinesInput struct {
	Project string `json:"project,omitempty" jsonschema:"project path to filter by"`
	Status  string `json:"status,omitempty" jsonschema:"pipeline status filter (running/pending/success/failed/canceled)"`
	Ref     string `json:"ref,omitempty" jsonschema:"git ref (branch/tag) to filter by"`
	Limit   int    `json:"limit,omitempty" jsonschema:"max number of pipelines to return"`
}

type ListJobsInput struct {
	ProjectID  int `json:"project_id" jsonschema:"GitLab project ID"`
	PipelineID int `json:"pipeline_id" jsonschema:"pipeline ID"`
}

type JobActionInput struct {
	ProjectID int `json:"project_id" jsonschema:"GitLab project ID"`
	JobID     int `json:"job_id" jsonschema:"job ID"`
}

type SearchProjectsInput struct {
	Query string `json:"query" jsonschema:"search query for project name or path"`
}

// Tool handlers

func listProjectsHandler(cfg *config.Config, pSvc *service.PipelineService) func(context.Context, *mcp.CallToolRequest, ListProjectsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ ListProjectsInput) (*mcp.CallToolResult, any, error) {
		projects, err := pSvc.LoadProjects(ctx, cfg.Projects)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(formatProjects(projects)), nil, nil
	}
}

func listPipelinesHandler(cfg *config.Config, pSvc *service.PipelineService) func(context.Context, *mcp.CallToolRequest, ListPipelinesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ListPipelinesInput) (*mcp.CallToolResult, any, error) {
		paths := cfg.Projects
		if input.Project != "" {
			paths = []string{input.Project}
		}
		limit := cfg.PipelineLimit
		if input.Limit > 0 {
			limit = input.Limit
		}

		pipelines, err := pSvc.LoadAllPipelines(ctx, paths, limit)
		if err != nil {
			return errResult(err), nil, nil
		}

		// Apply filters
		if input.Status != "" || input.Ref != "" {
			filtered := pipelines[:0]
			for _, p := range pipelines {
				if input.Status != "" && string(p.Status) != input.Status {
					continue
				}
				if input.Ref != "" && p.Ref != input.Ref {
					continue
				}
				filtered = append(filtered, p)
			}
			pipelines = filtered
		}

		return textResult(formatPipelines(pipelines)), nil, nil
	}
}

func listJobsHandler(pSvc *service.PipelineService) func(context.Context, *mcp.CallToolRequest, ListJobsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ListJobsInput) (*mcp.CallToolResult, any, error) {
		jobs, err := pSvc.ListJobs(ctx, input.ProjectID, input.PipelineID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(formatJobs(jobs)), nil, nil
	}
}

func getJobLogHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		rc, err := jSvc.GetJobLog(ctx, input.ProjectID, input.JobID)
		if err != nil {
			return errResult(err), nil, nil
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			return errResult(err), nil, nil
		}
		log := string(data)
		const maxLen = 50000
		if len(log) > maxLen {
			log = "... (truncated)\n" + log[len(log)-maxLen:]
		}
		return textResult(log), nil, nil
	}
}

func playJobHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		job, err := jSvc.PlayJob(ctx, input.ProjectID, input.JobID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Job started: %s", formatJob(*job))), nil, nil
	}
}

func retryJobHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		job, err := jSvc.RetryJob(ctx, input.ProjectID, input.JobID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Job retried: %s", formatJob(*job))), nil, nil
	}
}

func cancelJobHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		job, err := jSvc.CancelJob(ctx, input.ProjectID, input.JobID)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(fmt.Sprintf("Job canceled: %s", formatJob(*job))), nil, nil
	}
}

func searchProjectsHandler(pSvc *service.PipelineService) func(context.Context, *mcp.CallToolRequest, SearchProjectsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SearchProjectsInput) (*mcp.CallToolResult, any, error) {
		projects, err := pSvc.SearchProjects(ctx, input.Query)
		if err != nil {
			return errResult(err), nil, nil
		}
		return textResult(formatProjects(projects)), nil, nil
	}
}

// Helpers

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func errResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
		IsError: true,
	}
}
