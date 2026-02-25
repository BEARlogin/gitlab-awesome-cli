package gitlab

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	gogitlab "github.com/xanzy/go-gitlab"
)

type ProjectRepo struct {
	client *gogitlab.Client
}

func NewProjectRepo(client *gogitlab.Client) *ProjectRepo {
	return &ProjectRepo{client: client}
}

func (r *ProjectRepo) GetByPath(ctx context.Context, pathWithNS string) (*entity.Project, error) {
	p, _, err := r.client.Projects.GetProject(pathWithNS, nil, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return &entity.Project{
		ID:         p.ID,
		Name:       p.Name,
		PathWithNS: p.PathWithNamespace,
		WebURL:     p.WebURL,
	}, nil
}

func (r *ProjectRepo) Search(ctx context.Context, query string) ([]entity.Project, error) {
	opts := &gogitlab.ListProjectsOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 10},
		Search:      gogitlab.Ptr(query),
		OrderBy:     gogitlab.Ptr("name"),
	}
	projects, _, err := r.client.Projects.ListProjects(opts, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Project, len(projects))
	for i, p := range projects {
		result[i] = entity.Project{
			ID:         p.ID,
			Name:       p.Name,
			PathWithNS: p.PathWithNamespace,
			WebURL:     p.WebURL,
		}
	}
	return result, nil
}

func (r *ProjectRepo) ListPipelines(ctx context.Context, projectID int) ([]entity.Pipeline, error) {
	opts := &gogitlab.ListProjectPipelinesOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 20},
		OrderBy:     gogitlab.Ptr("id"),
		Sort:        gogitlab.Ptr("desc"),
	}
	pls, _, err := r.client.Pipelines.ListProjectPipelines(projectID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Pipeline, len(pls))
	for i, pl := range pls {
		result[i] = entity.Pipeline{
			ID:        pl.ID,
			ProjectID: projectID,
			Ref:       pl.Ref,
			Status:    valueobject.PipelineStatus(pl.Status),
			CreatedAt: *pl.CreatedAt,
		}
	}
	return result, nil
}
