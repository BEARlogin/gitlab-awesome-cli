package gitlab

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	gogitlab "github.com/xanzy/go-gitlab"
)

type PipelineRepo struct {
	client *gogitlab.Client
}

func NewPipelineRepo(client *gogitlab.Client) *PipelineRepo {
	return &PipelineRepo{client: client}
}

func (r *PipelineRepo) ListJobs(ctx context.Context, projectID, pipelineID int) ([]entity.Job, error) {
	log.Printf("[gitlab] ListJobs: project=%d pipeline=%d", projectID, pipelineID)
	opts := &gogitlab.ListJobsOptions{ListOptions: gogitlab.ListOptions{PerPage: 100}}

	// Regular jobs
	jobs, _, err := r.client.Jobs.ListPipelineJobs(projectID, pipelineID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ListJobs: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] ListJobs: got %d regular jobs", len(jobs))

	result := make([]entity.Job, 0, len(jobs))
	for _, j := range jobs {
		job := entity.Job{
			ID:         j.ID,
			PipelineID: pipelineID,
			ProjectID:  projectID,
			Name:       j.Name,
			Stage:      j.Stage,
			Status:     valueobject.JobStatus(j.Status),
			Duration:   j.Duration,
			WebURL:     j.WebURL,
		}
		if j.StartedAt != nil {
			job.StartedAt = j.StartedAt
		}
		if j.FinishedAt != nil {
			job.FinishedAt = j.FinishedAt
		}
		result = append(result, job)
	}

	// Bridge/trigger jobs
	bridges, _, err := r.client.Jobs.ListPipelineBridges(projectID, pipelineID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ListJobs: bridges error (non-fatal): %v", err)
	} else {
		log.Printf("[gitlab] ListJobs: got %d bridge jobs", len(bridges))
		for _, b := range bridges {
			job := entity.Job{
				ID:         b.ID,
				PipelineID: pipelineID,
				ProjectID:  projectID,
				Name:       b.Name,
				Stage:      b.Stage,
				Status:     valueobject.JobStatus(b.Status),
				Duration:   b.Duration,
				WebURL:     b.WebURL,
			}
			if b.StartedAt != nil {
				job.StartedAt = b.StartedAt
			}
			if b.FinishedAt != nil {
				job.FinishedAt = b.FinishedAt
			}
			result = append(result, job)
		}
	}

	log.Printf("[gitlab] ListJobs: total %d jobs", len(result))
	return result, nil
}

func (r *PipelineRepo) LoadAllPipelines(ctx context.Context, projectPaths []string, perProject int) ([]entity.Pipeline, error) {
	log.Printf("[gitlab] LoadAllPipelines: paths=%v perProject=%d", projectPaths, perProject)
	var all []entity.Pipeline
	for _, path := range projectPaths {
		p, _, err := r.client.Projects.GetProject(path, nil, gogitlab.WithContext(ctx))
		if err != nil {
			log.Printf("[gitlab] LoadAllPipelines: skip %s: %v", path, err)
			continue
		}

		opts := &gogitlab.ListProjectPipelinesOptions{
			ListOptions: gogitlab.ListOptions{PerPage: perProject},
			OrderBy:     gogitlab.Ptr("id"),
			Sort:        gogitlab.Ptr("desc"),
		}
		pls, _, err := r.client.Pipelines.ListProjectPipelines(p.ID, opts, gogitlab.WithContext(ctx))
		if err != nil {
			log.Printf("[gitlab] LoadAllPipelines: skip pipelines for %s: %v", path, err)
			continue
		}
		log.Printf("[gitlab] LoadAllPipelines: %s got %d pipelines", path, len(pls))

		for _, pl := range pls {
			createdAt := time.Time{}
			if pl.CreatedAt != nil {
				createdAt = *pl.CreatedAt
			}
			all = append(all, entity.Pipeline{
				ID:          pl.ID,
				ProjectID:   p.ID,
				ProjectPath: p.PathWithNamespace,
				Ref:         pl.Ref,
				Status:      valueobject.PipelineStatus(pl.Status),
				CreatedAt:   createdAt,
			})
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	return all, nil
}
