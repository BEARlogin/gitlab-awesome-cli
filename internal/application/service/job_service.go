package service

import (
	"context"
	"io"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/repository"
)

type JobService struct {
	jobRepo repository.JobRepository
}

func NewJobService(jr repository.JobRepository) *JobService {
	return &JobService{jobRepo: jr}
}

func (s *JobService) PlayJob(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	return s.jobRepo.Play(ctx, projectID, jobID)
}

func (s *JobService) RetryJob(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	return s.jobRepo.Retry(ctx, projectID, jobID)
}

func (s *JobService) CancelJob(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	return s.jobRepo.Cancel(ctx, projectID, jobID)
}

func (s *JobService) GetJobLog(ctx context.Context, projectID, jobID int) (io.ReadCloser, error) {
	return s.jobRepo.GetLog(ctx, projectID, jobID)
}
