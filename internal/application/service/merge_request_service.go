package service

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/repository"
)

type MergeRequestService struct {
	mrRepo     repository.MergeRequestRepository
	commitRepo repository.CommitRepository
}

func NewMergeRequestService(mr repository.MergeRequestRepository, cr repository.CommitRepository) *MergeRequestService {
	return &MergeRequestService{mrRepo: mr, commitRepo: cr}
}

func (s *MergeRequestService) ListMRs(ctx context.Context, projectID int, state string) ([]entity.MergeRequest, error) {
	return s.mrRepo.List(ctx, projectID, state)
}

func (s *MergeRequestService) GetMR(ctx context.Context, projectID, mrIID int) (*entity.MergeRequest, error) {
	return s.mrRepo.Get(ctx, projectID, mrIID)
}

func (s *MergeRequestService) ListNotes(ctx context.Context, projectID, mrIID int) ([]entity.MRNote, error) {
	return s.mrRepo.ListNotes(ctx, projectID, mrIID)
}

func (s *MergeRequestService) GetDiffs(ctx context.Context, projectID, mrIID int) ([]entity.MRDiff, error) {
	return s.mrRepo.GetDiffs(ctx, projectID, mrIID)
}

func (s *MergeRequestService) CreateMR(ctx context.Context, projectID int, opts entity.CreateMROptions) (*entity.MergeRequest, error) {
	return s.mrRepo.Create(ctx, projectID, opts)
}

func (s *MergeRequestService) ApproveMR(ctx context.Context, projectID, mrIID int) error {
	return s.mrRepo.Approve(ctx, projectID, mrIID)
}

func (s *MergeRequestService) MergeMR(ctx context.Context, projectID, mrIID int) (*entity.MergeRequest, error) {
	return s.mrRepo.Merge(ctx, projectID, mrIID)
}

func (s *MergeRequestService) ListCommits(ctx context.Context, projectID int, ref string) ([]entity.Commit, error) {
	return s.commitRepo.ListByRef(ctx, projectID, ref)
}
