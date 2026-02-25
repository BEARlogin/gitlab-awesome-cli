package gitlab

import (
	"context"

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
