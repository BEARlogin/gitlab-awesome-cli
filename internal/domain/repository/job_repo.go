package repository

import (
	"context"
	"io"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type JobRepository interface {
	Play(ctx context.Context, projectID, jobID int) (*entity.Job, error)
	Retry(ctx context.Context, projectID, jobID int) (*entity.Job, error)
	Cancel(ctx context.Context, projectID, jobID int) (*entity.Job, error)
	GetLog(ctx context.Context, projectID, jobID int) (io.ReadCloser, error)
}
