package gitlab

import (
	"bytes"
	"context"
	"io"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	gogitlab "github.com/xanzy/go-gitlab"
)

type JobRepo struct {
	client *gogitlab.Client
}

func NewJobRepo(client *gogitlab.Client) *JobRepo {
	return &JobRepo{client: client}
}

func (r *JobRepo) Play(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	j, _, err := r.client.Jobs.PlayJob(projectID, jobID, nil, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return mapJob(j, projectID), nil
}

func (r *JobRepo) Retry(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	j, _, err := r.client.Jobs.RetryJob(projectID, jobID, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return mapJob(j, projectID), nil
}

func (r *JobRepo) Cancel(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	j, _, err := r.client.Jobs.CancelJob(projectID, jobID, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return mapJob(j, projectID), nil
}

func (r *JobRepo) GetLog(ctx context.Context, projectID, jobID int) (io.ReadCloser, error) {
	trace, _, err := r.client.Jobs.GetTraceFile(projectID, jobID, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(trace)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func mapJob(j *gogitlab.Job, projectID int) *entity.Job {
	job := &entity.Job{
		ID:        j.ID,
		ProjectID: projectID,
		Name:      j.Name,
		Stage:     j.Stage,
		Status:    valueobject.JobStatus(j.Status),
		Duration:  j.Duration,
		WebURL:    j.WebURL,
	}
	if j.Pipeline.ID != 0 {
		job.PipelineID = j.Pipeline.ID
	}
	return job
}
