package mcp

import (
	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer(cfg *config.Config, pSvc *service.PipelineService, jSvc *service.JobService, mrSvc *service.MergeRequestService, version string) *mcp.Server {
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

	// Merge Request tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_merge_requests",
		Description: "List merge requests for a project",
	}, listMergeRequestsHandler(mrSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_merge_request",
		Description: "Get details of a specific merge request",
	}, getMergeRequestHandler(mrSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_mr_notes",
		Description: "List comments/notes on a merge request",
	}, listMRNotesHandler(mrSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_mr_diffs",
		Description: "Get diffs of a merge request",
	}, getMRDiffsHandler(mrSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "approve_mr",
		Description: "Approve a merge request",
	}, approveMRHandler(mrSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "merge_mr",
		Description: "Merge a merge request",
	}, mergeMRHandler(mrSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_merge_request",
		Description: "Create a new merge request",
	}, createMRHandler(mrSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_branches",
		Description: "List repository branches, optionally filtered by name",
	}, listBranchesHandler(pSvc))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pipeline_commits",
		Description: "List commits for a pipeline ref (branch/tag)",
	}, listPipelineCommitsHandler(mrSvc))

	return server
}
