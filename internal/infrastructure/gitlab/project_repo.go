package gitlab

import (
	"context"
	"log"

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
	log.Printf("[gitlab] GetByPath: %s", pathWithNS)
	p, _, err := r.client.Projects.GetProject(pathWithNS, nil, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] GetByPath: error: %v", err)
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
	log.Printf("[gitlab] Search: query=%q", query)
	opts := &gogitlab.ListProjectsOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 10},
		Search:      gogitlab.Ptr(query),
		OrderBy:     gogitlab.Ptr("name"),
	}
	projects, _, err := r.client.Projects.ListProjects(opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] Search: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] Search: found %d projects", len(projects))
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

func (r *ProjectRepo) ListBranches(ctx context.Context, projectID int, search string) ([]string, error) {
	log.Printf("[gitlab] ListBranches: project=%d search=%q", projectID, search)
	opts := &gogitlab.ListBranchesOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 20},
	}
	if search != "" {
		opts.Search = gogitlab.Ptr(search)
	}
	branches, _, err := r.client.Branches.ListBranches(projectID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ListBranches: error: %v", err)
		return nil, err
	}
	result := make([]string, len(branches))
	for i, b := range branches {
		result[i] = b.Name
	}
	log.Printf("[gitlab] ListBranches: got %d branches", len(result))
	return result, nil
}

func (r *ProjectRepo) ListPipelines(ctx context.Context, projectID int) ([]entity.Pipeline, error) {
	log.Printf("[gitlab] ListPipelines: project=%d", projectID)
	opts := &gogitlab.ListProjectPipelinesOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 20},
		OrderBy:     gogitlab.Ptr("id"),
		Sort:        gogitlab.Ptr("desc"),
	}
	pls, _, err := r.client.Pipelines.ListProjectPipelines(projectID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		log.Printf("[gitlab] ListPipelines: error: %v", err)
		return nil, err
	}
	log.Printf("[gitlab] ListPipelines: project=%d got %d pipelines", projectID, len(pls))
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
