package mcp

import (
	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer(cfg *config.Config, pSvc *service.PipelineService, jSvc *service.JobService, version string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "glcli-mcp",
		Version: version,
	}, nil)

	// Resources
	server.AddResource(&mcp.Resource{
		URI:         "gitlab://config",
		Name:        "GitLab Configuration",
		Description: "Current glcli configuration (without token)",
	}, configResourceHandler(cfg))

	// Tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List configured GitLab projects with pipeline counts",
	}, listProjectsHandler(cfg, pSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pipelines",
		Description: "List pipelines across configured projects. Optionally filter by project, status, ref, or limit.",
	}, listPipelinesHandler(cfg, pSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_jobs",
		Description: "List jobs for a specific pipeline",
	}, listJobsHandler(pSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_job_log",
		Description: "Get the log output of a specific job",
	}, getJobLogHandler(jSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "play_job",
		Description: "Start a manual job",
	}, playJobHandler(jSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "retry_job",
		Description: "Retry a failed job",
	}, retryJobHandler(jSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cancel_job",
		Description: "Cancel a running or pending job",
	}, cancelJobHandler(jSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_projects",
		Description: "Search GitLab projects by name or path",
	}, searchProjectsHandler(pSvc))

	return server
}
