package repository

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type CommitRepository interface {
	ListByRef(ctx context.Context, projectID int, ref string) ([]entity.Commit, error)
}
