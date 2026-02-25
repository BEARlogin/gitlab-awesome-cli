package service

import (
	"context"
	"sort"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/repository"
)

type PipelineService struct {
	projectRepo  repository.ProjectRepository
	pipelineRepo repository.PipelineRepository
}

func NewPipelineService(pr repository.ProjectRepository, plr repository.PipelineRepository) *PipelineService {
	return &PipelineService{projectRepo: pr, pipelineRepo: plr}
}

func (s *PipelineService) LoadProjects(ctx context.Context, paths []string) ([]entity.Project, error) {
	projects := make([]entity.Project, 0, len(paths))
	for _, path := range paths {
		p, err := s.projectRepo.GetByPath(ctx, path)
		if err != nil { return nil, err }
		pipelines, err := s.projectRepo.ListPipelines(ctx, p.ID)
		if err != nil { return nil, err }
		p.PipelineCount = len(pipelines)
		for _, pl := range pipelines {
			if pl.Status.IsActive() { p.ActiveCount++ }
		}
		projects = append(projects, *p)
	}
	return projects, nil
}

func (s *PipelineService) ListPipelines(ctx context.Context, projectID int) ([]entity.Pipeline, error) {
	return s.projectRepo.ListPipelines(ctx, projectID)
}

func (s *PipelineService) LoadAllPipelines(ctx context.Context, paths []string, limit int) ([]entity.Pipeline, error) {
	var all []entity.Pipeline
	for _, path := range paths {
		p, err := s.projectRepo.GetByPath(ctx, path)
		if err != nil {
			return nil, err
		}
		pls, err := s.projectRepo.ListPipelines(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		for i := range pls {
			pls[i].ProjectPath = p.PathWithNS
		}
		all = append(all, pls...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (s *PipelineService) SearchProjects(ctx context.Context, query string) ([]entity.Project, error) {
	return s.projectRepo.Search(ctx, query)
}

func (s *PipelineService) ListJobs(ctx context.Context, projectID, pipelineID int) ([]entity.Job, error) {
	return s.pipelineRepo.ListJobs(ctx, projectID, pipelineID)
}
