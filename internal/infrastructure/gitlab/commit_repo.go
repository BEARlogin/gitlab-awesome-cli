package gitlab

import (
	"context"
	"log"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	gogitlab "github.com/xanzy/go-gitlab"
)

type CommitRepo struct {
	client *gogitlab.Client
}

func NewCommitRepo(client *gogitlab.Client) *CommitRepo {
	return &CommitRepo{client: client}
}

func (r *CommitRepo) ListByRef(ctx context.Context, projectID int, ref string) ([]entity.Commit, error) {
	log.Printf("[gitlab] ListCommits: project=%d ref=%s", projectID, ref)
	opts := &gogitlab.ListCommitsOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 50},
		RefName:     gogitlab.Ptr(ref),
	}
	commits, _, err := r.client.Commits.ListCommits(projectID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ListCommits: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] ListCommits: got %d commits", len(commits))
	result := make([]entity.Commit, len(commits))
	for i, c := range commits {
		result[i] = entity.Commit{
			ShortID:     c.ShortID,
			Title:       c.Title,
			AuthorName:  c.AuthorName,
			AuthorEmail: c.AuthorEmail,
			WebURL:      c.WebURL,
		}
		if c.CreatedAt != nil {
			result[i].CreatedAt = *c.CreatedAt
		}
	}
	return result, nil
}
