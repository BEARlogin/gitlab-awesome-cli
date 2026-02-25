package gitlab

import (
	"context"
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
	opts := &gogitlab.ListJobsOptions{ListOptions: gogitlab.ListOptions{PerPage: 100}}
	jobs, _, err := r.client.Jobs.ListPipelineJobs(projectID, pipelineID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Job, len(jobs))
	for i, j := range jobs {
		result[i] = entity.Job{
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
			result[i].StartedAt = j.StartedAt
		}
		if j.FinishedAt != nil {
			result[i].FinishedAt = j.FinishedAt
		}
	}
	return result, nil
}

func (r *PipelineRepo) LoadAllPipelines(ctx context.Context, projectPaths []string, perProject int) ([]entity.Pipeline, error) {
	var all []entity.Pipeline
	for _, path := range projectPaths {
		p, _, err := r.client.Projects.GetProject(path, nil, gogitlab.WithContext(ctx))
		if err != nil {
			continue
		}

		opts := &gogitlab.ListProjectPipelinesOptions{
			ListOptions: gogitlab.ListOptions{PerPage: perProject},
			OrderBy:     gogitlab.Ptr("id"),
			Sort:        gogitlab.Ptr("desc"),
		}
		pls, _, err := r.client.Pipelines.ListProjectPipelines(p.ID, opts, gogitlab.WithContext(ctx))
		if err != nil {
			continue
		}

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
