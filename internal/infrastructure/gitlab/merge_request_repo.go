package gitlab

import (
	"context"
	"log"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	gogitlab "github.com/xanzy/go-gitlab"
)

type MergeRequestRepo struct {
	client *gogitlab.Client
}

func NewMergeRequestRepo(client *gogitlab.Client) *MergeRequestRepo {
	return &MergeRequestRepo{client: client}
}

func (r *MergeRequestRepo) List(ctx context.Context, projectID int, state string) ([]entity.MergeRequest, error) {
	log.Printf("[gitlab] ListMergeRequests: project=%d state=%s", projectID, state)
	opts := &gogitlab.ListProjectMergeRequestsOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 50},
		OrderBy:     gogitlab.Ptr("updated_at"),
		Sort:        gogitlab.Ptr("desc"),
	}
	if state != "" {
		opts.State = gogitlab.Ptr(state)
	}
	mrs, _, err := r.client.MergeRequests.ListProjectMergeRequests(projectID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ListMergeRequests: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] ListMergeRequests: got %d MRs", len(mrs))
	result := make([]entity.MergeRequest, len(mrs))
	for i, mr := range mrs {
		result[i] = mapMergeRequest(mr, projectID)
	}
	return result, nil
}

func (r *MergeRequestRepo) Get(ctx context.Context, projectID, mrIID int) (*entity.MergeRequest, error) {
	log.Printf("[gitlab] GetMergeRequest: project=%d mr=!%d", projectID, mrIID)
	mr, _, err := r.client.MergeRequests.GetMergeRequest(projectID, mrIID, nil, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] GetMergeRequest: error: %v", err)
		return nil, err
	}
	result := mapMergeRequest(mr, projectID)
	return &result, nil
}

func (r *MergeRequestRepo) ListNotes(ctx context.Context, projectID, mrIID int) ([]entity.MRNote, error) {
	log.Printf("[gitlab] ListMRNotes: project=%d mr=!%d", projectID, mrIID)
	opts := &gogitlab.ListMergeRequestNotesOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 100},
		OrderBy:     gogitlab.Ptr("created_at"),
		Sort:        gogitlab.Ptr("asc"),
	}
	notes, _, err := r.client.Notes.ListMergeRequestNotes(projectID, mrIID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ListMRNotes: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] ListMRNotes: got %d notes", len(notes))
	result := make([]entity.MRNote, len(notes))
	for i, n := range notes {
		result[i] = entity.MRNote{
			ID:        n.ID,
			Author:    n.Author.Username,
			Body:      n.Body,
			CreatedAt: *n.CreatedAt,
			System:    n.System,
		}
	}
	return result, nil
}

func (r *MergeRequestRepo) GetDiffs(ctx context.Context, projectID, mrIID int) ([]entity.MRDiff, error) {
	log.Printf("[gitlab] GetMRDiffs: project=%d mr=!%d", projectID, mrIID)
	opts := &gogitlab.ListMergeRequestDiffsOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 100},
	}
	diffs, _, err := r.client.MergeRequests.ListMergeRequestDiffs(projectID, mrIID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] GetMRDiffs: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] GetMRDiffs: got %d diffs", len(diffs))
	result := make([]entity.MRDiff, len(diffs))
	for i, d := range diffs {
		result[i] = entity.MRDiff{
			OldPath:     d.OldPath,
			NewPath:     d.NewPath,
			Diff:        d.Diff,
			NewFile:     d.NewFile,
			DeletedFile: d.DeletedFile,
			RenamedFile: d.RenamedFile,
		}
	}
	return result, nil
}

func (r *MergeRequestRepo) Create(ctx context.Context, projectID int, opts entity.CreateMROptions) (*entity.MergeRequest, error) {
	log.Printf("[gitlab] CreateMR: project=%d source=%s target=%s title=%q draft=%v", projectID, opts.SourceBranch, opts.TargetBranch, opts.Title, opts.Draft)
	title := opts.Title
	if opts.Draft {
		title = "Draft: " + title
	}
	createOpts := &gogitlab.CreateMergeRequestOptions{
		SourceBranch: gogitlab.Ptr(opts.SourceBranch),
		TargetBranch: gogitlab.Ptr(opts.TargetBranch),
		Title:        gogitlab.Ptr(title),
	}
	if opts.Description != "" {
		createOpts.Description = gogitlab.Ptr(opts.Description)
	}
	mr, _, err := r.client.MergeRequests.CreateMergeRequest(projectID, createOpts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] CreateMR: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] CreateMR: ok, iid=%d", mr.IID)
	result := mapMergeRequest(mr, projectID)
	return &result, nil
}

func (r *MergeRequestRepo) Approve(ctx context.Context, projectID, mrIID int) error {
	log.Printf("[gitlab] ApproveMR: project=%d mr=!%d", projectID, mrIID)
	_, _, err := r.client.MergeRequestApprovals.ApproveMergeRequest(projectID, mrIID, nil, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ApproveMR: error: %v", err)
		return err
	}
	log.Printf("[gitlab] ApproveMR: ok")
	return nil
}

func (r *MergeRequestRepo) Merge(ctx context.Context, projectID, mrIID int) (*entity.MergeRequest, error) {
	log.Printf("[gitlab] MergeMR: project=%d mr=!%d", projectID, mrIID)
	mr, _, err := r.client.MergeRequests.AcceptMergeRequest(projectID, mrIID, nil, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] MergeMR: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] MergeMR: ok, state=%s", mr.State)
	result := mapMergeRequest(mr, projectID)
	return &result, nil
}

func mapMergeRequest(mr *gogitlab.MergeRequest, projectID int) entity.MergeRequest {
	m := entity.MergeRequest{
		ID:           mr.ID,
		IID:          mr.IID,
		ProjectID:    projectID,
		Title:        mr.Title,
		Description:  mr.Description,
		State:        mr.State,
		SourceBranch: mr.SourceBranch,
		TargetBranch: mr.TargetBranch,
		MergeStatus:  mr.MergeStatus,
		Draft:        mr.Draft,
		WebURL:       mr.WebURL,
	}
	if mr.Author != nil {
		m.Author = mr.Author.Username
	}
	if mr.CreatedAt != nil {
		m.CreatedAt = *mr.CreatedAt
	}
	if mr.UpdatedAt != nil {
		m.UpdatedAt = *mr.UpdatedAt
	}
	return m
}
