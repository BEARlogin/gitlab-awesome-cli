package repository

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type PipelineRepository interface {
	ListJobs(ctx context.Context, projectID, pipelineID int) ([]entity.Job, error)
	LoadAllPipelines(ctx context.Context, projectPaths []string, perProject int) ([]entity.Pipeline, error)
}
