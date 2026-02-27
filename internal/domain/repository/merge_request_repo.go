package repository

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type MergeRequestRepository interface {
	List(ctx context.Context, projectID int, state string) ([]entity.MergeRequest, error)
	Get(ctx context.Context, projectID, mrIID int) (*entity.MergeRequest, error)
	ListNotes(ctx context.Context, projectID, mrIID int) ([]entity.MRNote, error)
	GetDiffs(ctx context.Context, projectID, mrIID int) ([]entity.MRDiff, error)
	Create(ctx context.Context, projectID int, opts entity.CreateMROptions) (*entity.MergeRequest, error)
	Approve(ctx context.Context, projectID, mrIID int) error
	Merge(ctx context.Context, projectID, mrIID int) (*entity.MergeRequest, error)
}
