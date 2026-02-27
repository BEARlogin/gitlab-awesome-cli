package mcp

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
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

type ListMergeRequestsInput struct {
	ProjectID int    `json:"project_id" jsonschema:"GitLab project ID"`
	State     string `json:"state,omitempty" jsonschema:"MR state filter (opened/merged/closed)"`
}

type MRInput struct {
	ProjectID int `json:"project_id" jsonschema:"GitLab project ID"`
	MRIID     int `json:"mr_iid" jsonschema:"merge request IID (project-scoped ID)"`
}

type CreateMRInput struct {
	ProjectID    int    `json:"project_id" jsonschema:"GitLab project ID"`
	SourceBranch string `json:"source_branch" jsonschema:"source branch name"`
	TargetBranch string `json:"target_branch" jsonschema:"target branch name"`
	Title        string `json:"title" jsonschema:"merge request title"`
	Description  string `json:"description,omitempty" jsonschema:"merge request description"`
	Draft        bool   `json:"draft,omitempty" jsonschema:"create as draft MR"`
}

type ListPipelineCommitsInput struct {
	ProjectID int    `json:"project_id" jsonschema:"GitLab project ID"`
	Ref       string `json:"ref" jsonschema:"git ref (branch/tag) to list commits for"`
}

// Tool handlers

func listProjectsHandler(cfg *config.Config, pSvc *service.PipelineService) func(context.Context, *mcp.CallToolRequest, ListProjectsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ ListProjectsInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] list_projects: paths=%v", cfg.Projects)
		projects, err := pSvc.LoadProjects(ctx, cfg.Projects)
		if err != nil {
			log.Printf("[tool] list_projects: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] list_projects: ok, %d projects", len(projects))
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
		log.Printf("[tool] list_pipelines: paths=%v status=%q ref=%q limit=%d", paths, input.Status, input.Ref, limit)

		pipelines, err := pSvc.LoadAllPipelines(ctx, paths, limit)
		if err != nil {
			log.Printf("[tool] list_pipelines: error: %v", err)
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

		log.Printf("[tool] list_pipelines: ok, %d pipelines", len(pipelines))
		return textResult(formatPipelines(pipelines)), nil, nil
	}
}

func listJobsHandler(pSvc *service.PipelineService) func(context.Context, *mcp.CallToolRequest, ListJobsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ListJobsInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] list_jobs: project=%d pipeline=%d", input.ProjectID, input.PipelineID)
		jobs, err := pSvc.ListJobs(ctx, input.ProjectID, input.PipelineID)
		if err != nil {
			log.Printf("[tool] list_jobs: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] list_jobs: ok, %d jobs", len(jobs))
		return textResult(formatJobs(jobs)), nil, nil
	}
}

func getJobLogHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] get_job_log: project=%d job=%d", input.ProjectID, input.JobID)
		rc, err := jSvc.GetJobLog(ctx, input.ProjectID, input.JobID)
		if err != nil {
			log.Printf("[tool] get_job_log: error: %v", err)
			return errResult(err), nil, nil
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			log.Printf("[tool] get_job_log: read error: %v", err)
			return errResult(err), nil, nil
		}
		jobLog := string(data)
		const maxLen = 50000
		if len(jobLog) > maxLen {
			log.Printf("[tool] get_job_log: truncating %d -> %d bytes", len(jobLog), maxLen)
			jobLog = "... (truncated)\n" + jobLog[len(jobLog)-maxLen:]
		}
		log.Printf("[tool] get_job_log: ok, %d bytes", len(jobLog))
		return textResult(jobLog), nil, nil
	}
}

func playJobHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] play_job: project=%d job=%d", input.ProjectID, input.JobID)
		job, err := jSvc.PlayJob(ctx, input.ProjectID, input.JobID)
		if err != nil {
			log.Printf("[tool] play_job: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] play_job: ok, status=%s", job.Status)
		return textResult(fmt.Sprintf("Job started: %s", formatJob(*job))), nil, nil
	}
}

func retryJobHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] retry_job: project=%d job=%d", input.ProjectID, input.JobID)
		job, err := jSvc.RetryJob(ctx, input.ProjectID, input.JobID)
		if err != nil {
			log.Printf("[tool] retry_job: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] retry_job: ok, status=%s", job.Status)
		return textResult(fmt.Sprintf("Job retried: %s", formatJob(*job))), nil, nil
	}
}

func cancelJobHandler(jSvc *service.JobService) func(context.Context, *mcp.CallToolRequest, JobActionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input JobActionInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] cancel_job: project=%d job=%d", input.ProjectID, input.JobID)
		job, err := jSvc.CancelJob(ctx, input.ProjectID, input.JobID)
		if err != nil {
			log.Printf("[tool] cancel_job: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] cancel_job: ok, status=%s", job.Status)
		return textResult(fmt.Sprintf("Job canceled: %s", formatJob(*job))), nil, nil
	}
}

func searchProjectsHandler(pSvc *service.PipelineService) func(context.Context, *mcp.CallToolRequest, SearchProjectsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input SearchProjectsInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] search_projects: query=%q", input.Query)
		projects, err := pSvc.SearchProjects(ctx, input.Query)
		if err != nil {
			log.Printf("[tool] search_projects: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] search_projects: ok, %d results", len(projects))
		return textResult(formatProjects(projects)), nil, nil
	}
}

// Merge Request handlers

func listMergeRequestsHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, ListMergeRequestsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ListMergeRequestsInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] list_merge_requests: project=%d state=%q", input.ProjectID, input.State)
		mrs, err := mrSvc.ListMRs(ctx, input.ProjectID, input.State)
		if err != nil {
			log.Printf("[tool] list_merge_requests: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] list_merge_requests: ok, %d MRs", len(mrs))
		return textResult(formatMergeRequests(mrs)), nil, nil
	}
}

func getMergeRequestHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, MRInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input MRInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] get_merge_request: project=%d mr=!%d", input.ProjectID, input.MRIID)
		mr, err := mrSvc.GetMR(ctx, input.ProjectID, input.MRIID)
		if err != nil {
			log.Printf("[tool] get_merge_request: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] get_merge_request: ok")
		return textResult(formatMergeRequestDetail(mr)), nil, nil
	}
}

func listMRNotesHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, MRInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input MRInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] list_mr_notes: project=%d mr=!%d", input.ProjectID, input.MRIID)
		notes, err := mrSvc.ListNotes(ctx, input.ProjectID, input.MRIID)
		if err != nil {
			log.Printf("[tool] list_mr_notes: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] list_mr_notes: ok, %d notes", len(notes))
		return textResult(formatMRNotes(notes)), nil, nil
	}
}

func getMRDiffsHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, MRInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input MRInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] get_mr_diffs: project=%d mr=!%d", input.ProjectID, input.MRIID)
		diffs, err := mrSvc.GetDiffs(ctx, input.ProjectID, input.MRIID)
		if err != nil {
			log.Printf("[tool] get_mr_diffs: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] get_mr_diffs: ok, %d files", len(diffs))
		return textResult(formatMRDiffs(diffs)), nil, nil
	}
}

func approveMRHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, MRInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input MRInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] approve_mr: project=%d mr=!%d", input.ProjectID, input.MRIID)
		err := mrSvc.ApproveMR(ctx, input.ProjectID, input.MRIID)
		if err != nil {
			log.Printf("[tool] approve_mr: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] approve_mr: ok")
		return textResult(fmt.Sprintf("Merge request !%d approved.", input.MRIID)), nil, nil
	}
}

func mergeMRHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, MRInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input MRInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] merge_mr: project=%d mr=!%d", input.ProjectID, input.MRIID)
		mr, err := mrSvc.MergeMR(ctx, input.ProjectID, input.MRIID)
		if err != nil {
			log.Printf("[tool] merge_mr: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] merge_mr: ok, state=%s", mr.State)
		return textResult(fmt.Sprintf("Merge request merged: %s", formatMergeRequest(*mr))), nil, nil
	}
}

func createMRHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, CreateMRInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input CreateMRInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] create_merge_request: project=%d source=%s target=%s title=%q", input.ProjectID, input.SourceBranch, input.TargetBranch, input.Title)
		opts := entity.CreateMROptions{
			SourceBranch: input.SourceBranch,
			TargetBranch: input.TargetBranch,
			Title:        input.Title,
			Description:  input.Description,
			Draft:        input.Draft,
		}
		mr, err := mrSvc.CreateMR(ctx, input.ProjectID, opts)
		if err != nil {
			log.Printf("[tool] create_merge_request: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] create_merge_request: ok, iid=%d", mr.IID)
		return textResult(fmt.Sprintf("Merge request created: %s", formatMergeRequest(*mr))), nil, nil
	}
}

func listPipelineCommitsHandler(mrSvc *service.MergeRequestService) func(context.Context, *mcp.CallToolRequest, ListPipelineCommitsInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ListPipelineCommitsInput) (*mcp.CallToolResult, any, error) {
		log.Printf("[tool] list_pipeline_commits: project=%d ref=%s", input.ProjectID, input.Ref)
		commits, err := mrSvc.ListCommits(ctx, input.ProjectID, input.Ref)
		if err != nil {
			log.Printf("[tool] list_pipeline_commits: error: %v", err)
			return errResult(err), nil, nil
		}
		log.Printf("[tool] list_pipeline_commits: ok, %d commits", len(commits))
		return textResult(formatCommits(commits)), nil, nil
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
