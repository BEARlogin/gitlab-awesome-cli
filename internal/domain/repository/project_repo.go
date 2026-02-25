package repository

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type ProjectRepository interface {
	GetByPath(ctx context.Context, pathWithNS string) (*entity.Project, error)
	ListPipelines(ctx context.Context, projectID int) ([]entity.Pipeline, error)
}
