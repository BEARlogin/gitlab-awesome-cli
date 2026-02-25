package mcp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func configResourceHandler(cfg *config.Config) func(context.Context, *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return func(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		log.Print("[resource] gitlab://config: read")
		var b strings.Builder
		fmt.Fprintf(&b, "GitLab URL: %s\n", cfg.GitLabURL)
		fmt.Fprintf(&b, "Projects: %s\n", strings.Join(cfg.Projects, ", "))
		fmt.Fprintf(&b, "Refresh Interval: %s\n", cfg.RefreshInterval)
		fmt.Fprintf(&b, "Pipeline Limit: %d\n", cfg.PipelineLimit)

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:  "gitlab://config",
				Text: b.String(),
			}},
		}, nil
	}
}
