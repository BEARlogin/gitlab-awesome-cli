# GitLab Awesome CLI — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Interactive TUI for GitLab (like k9s) with multi-view navigation: projects, pipelines, jobs, logs.

**Architecture:** DDD with 4 layers — Domain (entities, repository interfaces), Application (use cases), Infrastructure (GitLab API, config), Presentation (Bubble Tea TUI). Clean dependency rule: inner layers never import outer layers.

**Tech Stack:** Go 1.22+, charmbracelet/bubbletea, charmbracelet/bubbles, charmbracelet/lipgloss, xanzy/go-gitlab, gopkg.in/yaml.v3

---

## DDD Structure

```
cmd/glcli/main.go                         — entrypoint
internal/
  domain/
    entity/
      project.go                           — Project entity
      pipeline.go                          — Pipeline entity
      job.go                               — Job entity
    valueobject/
      status.go                            — PipelineStatus, JobStatus value objects
    repository/
      project_repo.go                      — ProjectRepository interface
      pipeline_repo.go                     — PipelineRepository interface
      job_repo.go                          — JobRepository interface
  application/
    service/
      pipeline_service.go                  — PipelineService (list, refresh)
      job_service.go                       — JobService (run, retry, cancel, log stream)
  infrastructure/
    gitlab/
      client.go                            — GitLab API client factory
      project_repo.go                      — ProjectRepository implementation
      pipeline_repo.go                     — PipelineRepository implementation
      job_repo.go                          — JobRepository implementation
    config/
      config.go                            — Config loading/saving, setup wizard
  presentation/
    tui/
      app.go                               — Root Bubble Tea model, view router
      views/
        projects.go                        — Projects list view
        pipelines.go                       — Pipelines list view
        jobs.go                            — Jobs list view
        log.go                             — Job log streaming view
      components/
        breadcrumb.go                      — Navigation breadcrumb
        statusbar.go                       — Hotkey hints
        confirm.go                         — Confirmation dialog
      styles/
        theme.go                           — Lipgloss styles & colors
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/glcli/main.go`
- Create: `.gitignore`

**Step 1: Initialize Go module**

Run: `go mod init github.com/bearlogin/gitlab-awesome-cli`

**Step 2: Create .gitignore**

```
/glcli
/dist/
.DS_Store
```

**Step 3: Create minimal main.go**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("glcli - GitLab Awesome CLI")
	os.Exit(0)
}
```

**Step 4: Verify it compiles**

Run: `go run cmd/glcli/main.go`
Expected: prints "glcli - GitLab Awesome CLI"

**Step 5: Init git and commit**

```bash
git init
git add .
git commit -m "feat: initial project scaffolding"
```

---

### Task 2: Domain Layer — Entities & Value Objects

**Files:**
- Create: `internal/domain/entity/project.go`
- Create: `internal/domain/entity/pipeline.go`
- Create: `internal/domain/entity/job.go`
- Create: `internal/domain/valueobject/status.go`

**Step 1: Create status value objects**

```go
// internal/domain/valueobject/status.go
package valueobject

type PipelineStatus string

const (
	PipelineRunning  PipelineStatus = "running"
	PipelinePending  PipelineStatus = "pending"
	PipelineSuccess  PipelineStatus = "success"
	PipelineFailed   PipelineStatus = "failed"
	PipelineCanceled PipelineStatus = "canceled"
	PipelineSkipped  PipelineStatus = "skipped"
	PipelineManual   PipelineStatus = "manual"
	PipelineCreated  PipelineStatus = "created"
)

func (s PipelineStatus) Symbol() string {
	switch s {
	case PipelineRunning:
		return "●"
	case PipelinePending, PipelineCreated:
		return "◌"
	case PipelineSuccess:
		return "✓"
	case PipelineFailed:
		return "✗"
	case PipelineCanceled:
		return "⊘"
	case PipelineSkipped:
		return "»"
	case PipelineManual:
		return "⏸"
	default:
		return "?"
	}
}

func (s PipelineStatus) IsActive() bool {
	return s == PipelineRunning || s == PipelinePending
}

type JobStatus string

const (
	JobRunning  JobStatus = "running"
	JobPending  JobStatus = "pending"
	JobSuccess  JobStatus = "success"
	JobFailed   JobStatus = "failed"
	JobCanceled JobStatus = "canceled"
	JobSkipped  JobStatus = "skipped"
	JobManual   JobStatus = "manual"
	JobCreated  JobStatus = "created"
)

func (s JobStatus) Symbol() string {
	switch s {
	case JobRunning:
		return "●"
	case JobPending, JobCreated:
		return "◌"
	case JobSuccess:
		return "✓"
	case JobFailed:
		return "✗"
	case JobCanceled:
		return "⊘"
	case JobSkipped:
		return "»"
	case JobManual:
		return "⏸"
	default:
		return "?"
	}
}

func (s JobStatus) IsActionable() bool {
	return s == JobManual || s == JobFailed
}

func (s JobStatus) CanCancel() bool {
	return s == JobRunning || s == JobPending
}
```

**Step 2: Create Project entity**

```go
// internal/domain/entity/project.go
package entity

type Project struct {
	ID            int
	Name          string
	PathWithNS    string
	WebURL        string
	PipelineCount int
	ActiveCount   int
}
```

**Step 3: Create Pipeline entity**

```go
// internal/domain/entity/pipeline.go
package entity

import (
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
)

type Pipeline struct {
	ID        int
	ProjectID int
	Ref       string
	Status    valueobject.PipelineStatus
	CreatedAt time.Time
	Duration  int
	JobCount  int
}
```

**Step 4: Create Job entity**

```go
// internal/domain/entity/job.go
package entity

import (
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
)

type Job struct {
	ID         int
	PipelineID int
	ProjectID  int
	Name       string
	Stage      string
	Status     valueobject.JobStatus
	Duration   float64
	StartedAt  *time.Time
	FinishedAt *time.Time
	WebURL     string
}
```

**Step 5: Verify compilation**

Run: `go build ./...`
Expected: no errors

**Step 6: Commit**

```bash
git add internal/domain/
git commit -m "feat: domain layer — entities and value objects"
```

---

### Task 3: Domain Layer — Repository Interfaces

**Files:**
- Create: `internal/domain/repository/project_repo.go`
- Create: `internal/domain/repository/pipeline_repo.go`
- Create: `internal/domain/repository/job_repo.go`

**Step 1: Create repository interfaces**

```go
// internal/domain/repository/project_repo.go
package repository

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type ProjectRepository interface {
	GetByPath(ctx context.Context, pathWithNS string) (*entity.Project, error)
	ListPipelines(ctx context.Context, projectID int) ([]entity.Pipeline, error)
}
```

```go
// internal/domain/repository/pipeline_repo.go
package repository

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type PipelineRepository interface {
	ListJobs(ctx context.Context, projectID, pipelineID int) ([]entity.Job, error)
}
```

```go
// internal/domain/repository/job_repo.go
package repository

import (
	"context"
	"io"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
)

type JobRepository interface {
	Play(ctx context.Context, projectID, jobID int) (*entity.Job, error)
	Retry(ctx context.Context, projectID, jobID int) (*entity.Job, error)
	Cancel(ctx context.Context, projectID, jobID int) (*entity.Job, error)
	GetLog(ctx context.Context, projectID, jobID int) (io.ReadCloser, error)
}
```

**Step 2: Verify compilation**

Run: `go build ./...`

**Step 3: Commit**

```bash
git add internal/domain/repository/
git commit -m "feat: domain layer — repository interfaces"
```

---

### Task 4: Infrastructure — Config

**Files:**
- Create: `internal/infrastructure/config/config.go`

**Step 1: Install yaml dependency**

Run: `go get gopkg.in/yaml.v3`

**Step 2: Create config**

```go
// internal/infrastructure/config/config.go
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitLabURL       string        `yaml:"gitlab_url"`
	Token           string        `yaml:"token"`
	Projects        []string      `yaml:"projects"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".glcli.yaml")
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if cfg.RefreshInterval == 0 {
		cfg.RefreshInterval = 5 * time.Second
	}
	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func RunSetupWizard() (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	cfg := &Config{
		RefreshInterval: 5 * time.Second,
	}

	fmt.Print("GitLab URL (e.g. https://gitlab.example.com): ")
	url, _ := reader.ReadString('\n')
	cfg.GitLabURL = strings.TrimSpace(url)

	fmt.Print("Personal Access Token: ")
	token, _ := reader.ReadString('\n')
	cfg.Token = strings.TrimSpace(token)

	fmt.Print("Projects (comma-separated, e.g. group/project1,group/project2): ")
	projects, _ := reader.ReadString('\n')
	for _, p := range strings.Split(strings.TrimSpace(projects), ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			cfg.Projects = append(cfg.Projects, p)
		}
	}

	path := DefaultPath()
	if err := cfg.Save(path); err != nil {
		return nil, fmt.Errorf("saving config: %w", err)
	}
	fmt.Printf("Config saved to %s\n", path)
	return cfg, nil
}
```

**Step 3: Verify compilation**

Run: `go build ./...`

**Step 4: Commit**

```bash
git add internal/infrastructure/config/ go.mod go.sum
git commit -m "feat: infrastructure — config loading and setup wizard"
```

---

### Task 5: Infrastructure — GitLab Repository Implementations

**Files:**
- Create: `internal/infrastructure/gitlab/client.go`
- Create: `internal/infrastructure/gitlab/project_repo.go`
- Create: `internal/infrastructure/gitlab/pipeline_repo.go`
- Create: `internal/infrastructure/gitlab/job_repo.go`

**Step 1: Install go-gitlab**

Run: `go get github.com/xanzy/go-gitlab`

**Step 2: Create client factory**

```go
// internal/infrastructure/gitlab/client.go
package gitlab

import (
	"fmt"

	gogitlab "github.com/xanzy/go-gitlab"
)

func NewClient(baseURL, token string) (*gogitlab.Client, error) {
	client, err := gogitlab.NewClient(token, gogitlab.WithBaseURL(baseURL+"/api/v4"))
	if err != nil {
		return nil, fmt.Errorf("creating gitlab client: %w", err)
	}
	return client, nil
}
```

**Step 3: Create ProjectRepository implementation**

```go
// internal/infrastructure/gitlab/project_repo.go
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
```

**Step 4: Create PipelineRepository implementation**

```go
// internal/infrastructure/gitlab/pipeline_repo.go
package gitlab

import (
	"context"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	gogitlab "github.com/xanzy/go-gitlab"
)

type PipelineRepo struct {
	client *gogitlab.Client
}

func NewPipelineRepo(client *gogitlab.Client) *PipelineRepo {
	return &PipelineRepo{client: client}
}

func (r *PipelineRepo) ListJobs(ctx context.Context, projectID, pipelineID int) ([]entity.Job, error) {
	opts := &gogitlab.ListJobsOptions{
		ListOptions: gogitlab.ListOptions{PerPage: 100},
	}
	jobs, _, err := r.client.Jobs.ListPipelineJobs(projectID, pipelineID, opts, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	result := make([]entity.Job, len(jobs))
	for i, j := range jobs {
		result[i] = entity.Job{
			ID:         j.ID,
			PipelineID: pipelineID,
			ProjectID:  projectID,
			Name:       j.Name,
			Stage:      j.Stage,
			Status:     valueobject.JobStatus(j.Status),
			Duration:   j.Duration,
			WebURL:     j.WebURL,
		}
		if j.StartedAt != nil {
			result[i].StartedAt = j.StartedAt
		}
		if j.FinishedAt != nil {
			result[i].FinishedAt = j.FinishedAt
		}
	}
	return result, nil
}
```

**Step 5: Create JobRepository implementation**

```go
// internal/infrastructure/gitlab/job_repo.go
package gitlab

import (
	"bytes"
	"context"
	"io"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	gogitlab "github.com/xanzy/go-gitlab"
)

type JobRepo struct {
	client *gogitlab.Client
}

func NewJobRepo(client *gogitlab.Client) *JobRepo {
	return &JobRepo{client: client}
}

func (r *JobRepo) Play(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	j, _, err := r.client.Jobs.PlayJob(projectID, jobID, nil, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return mapJob(j, projectID), nil
}

func (r *JobRepo) Retry(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	j, _, err := r.client.Jobs.RetryJob(projectID, jobID, nil, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return mapJob(j, projectID), nil
}

func (r *JobRepo) Cancel(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	j, _, err := r.client.Jobs.CancelJob(projectID, jobID, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return mapJob(j, projectID), nil
}

func (r *JobRepo) GetLog(ctx context.Context, projectID, jobID int) (io.ReadCloser, error) {
	trace, _, err := r.client.Jobs.GetTraceFile(projectID, jobID, gogitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(trace)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func mapJob(j *gogitlab.Job, projectID int) *entity.Job {
	job := &entity.Job{
		ID:        j.ID,
		ProjectID: projectID,
		Name:      j.Name,
		Stage:     j.Stage,
		Status:    valueobject.JobStatus(j.Status),
		Duration:  j.Duration,
		WebURL:    j.WebURL,
	}
	if j.Pipeline.ID != 0 {
		job.PipelineID = j.Pipeline.ID
	}
	return job
}
```

**Step 6: Verify compilation**

Run: `go build ./...`

**Step 7: Commit**

```bash
git add internal/infrastructure/gitlab/ go.mod go.sum
git commit -m "feat: infrastructure — GitLab API repository implementations"
```

---

### Task 6: Application Layer — Services

**Files:**
- Create: `internal/application/service/pipeline_service.go`
- Create: `internal/application/service/job_service.go`

**Step 1: Create PipelineService**

```go
// internal/application/service/pipeline_service.go
package service

import (
	"context"

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
		if err != nil {
			return nil, err
		}
		pipelines, err := s.projectRepo.ListPipelines(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		p.PipelineCount = len(pipelines)
		for _, pl := range pipelines {
			if pl.Status.IsActive() {
				p.ActiveCount++
			}
		}
		projects = append(projects, *p)
	}
	return projects, nil
}

func (s *PipelineService) ListPipelines(ctx context.Context, projectID int) ([]entity.Pipeline, error) {
	return s.projectRepo.ListPipelines(ctx, projectID)
}

func (s *PipelineService) ListJobs(ctx context.Context, projectID, pipelineID int) ([]entity.Job, error) {
	return s.pipelineRepo.ListJobs(ctx, projectID, pipelineID)
}
```

**Step 2: Create JobService**

```go
// internal/application/service/job_service.go
package service

import (
	"context"
	"io"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/repository"
)

type JobService struct {
	jobRepo repository.JobRepository
}

func NewJobService(jr repository.JobRepository) *JobService {
	return &JobService{jobRepo: jr}
}

func (s *JobService) PlayJob(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	return s.jobRepo.Play(ctx, projectID, jobID)
}

func (s *JobService) RetryJob(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	return s.jobRepo.Retry(ctx, projectID, jobID)
}

func (s *JobService) CancelJob(ctx context.Context, projectID, jobID int) (*entity.Job, error) {
	return s.jobRepo.Cancel(ctx, projectID, jobID)
}

func (s *JobService) GetJobLog(ctx context.Context, projectID, jobID int) (io.ReadCloser, error) {
	return s.jobRepo.GetLog(ctx, projectID, jobID)
}
```

**Step 3: Verify compilation**

Run: `go build ./...`

**Step 4: Commit**

```bash
git add internal/application/
git commit -m "feat: application layer — pipeline and job services"
```

---

### Task 7: Presentation — Styles & Components

**Files:**
- Create: `internal/presentation/tui/styles/theme.go`
- Create: `internal/presentation/tui/components/breadcrumb.go`
- Create: `internal/presentation/tui/components/statusbar.go`
- Create: `internal/presentation/tui/components/confirm.go`

**Step 1: Create theme**

```go
// internal/presentation/tui/styles/theme.go
package styles

import "github.com/charmbracelet/lipgloss"

var (
	Purple    = lipgloss.Color("99")
	Green     = lipgloss.Color("82")
	Red       = lipgloss.Color("196")
	Yellow    = lipgloss.Color("214")
	Gray      = lipgloss.Color("245")
	LightGray = lipgloss.Color("241")
	Cyan      = lipgloss.Color("87")
	White     = lipgloss.Color("255")

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Purple).
		Padding(0, 1)

	ActiveTab = lipgloss.NewStyle().
		Bold(true).
		Foreground(White).
		Background(Purple).
		Padding(0, 2)

	InactiveTab = lipgloss.NewStyle().
		Foreground(Gray).
		Padding(0, 2)

	StatusSuccess = lipgloss.NewStyle().Foreground(Green)
	StatusFailed  = lipgloss.NewStyle().Foreground(Red)
	StatusRunning = lipgloss.NewStyle().Foreground(Cyan)
	StatusManual  = lipgloss.NewStyle().Foreground(Yellow)
	StatusPending = lipgloss.NewStyle().Foreground(Gray)

	Selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(White).
		Background(lipgloss.Color("62"))

	HelpKey = lipgloss.NewStyle().
		Bold(true).
		Foreground(Purple)

	HelpDesc = lipgloss.NewStyle().
		Foreground(Gray)
)
```

**Step 2: Create breadcrumb component**

```go
// internal/presentation/tui/components/breadcrumb.go
package components

import (
	"strings"

	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type Breadcrumb struct {
	Parts []string
}

func NewBreadcrumb() Breadcrumb {
	return Breadcrumb{}
}

func (b Breadcrumb) View() string {
	if len(b.Parts) == 0 {
		return styles.Title.Render("glcli")
	}
	return styles.Title.Render("glcli > " + strings.Join(b.Parts, " > "))
}
```

**Step 3: Create statusbar component**

```go
// internal/presentation/tui/components/statusbar.go
package components

import (
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type HotkeyHint struct {
	Key  string
	Desc string
}

type StatusBar struct {
	Hints []HotkeyHint
}

func NewStatusBar(hints []HotkeyHint) StatusBar {
	return StatusBar{Hints: hints}
}

func (s StatusBar) View() string {
	var result string
	for i, h := range s.Hints {
		if i > 0 {
			result += "  "
		}
		result += styles.HelpKey.Render(h.Key) + " " + styles.HelpDesc.Render(h.Desc)
	}
	return result
}
```

**Step 4: Create confirm dialog**

```go
// internal/presentation/tui/components/confirm.go
package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type ConfirmResult struct {
	Confirmed bool
	Action    string
	JobID     int
	ProjectID int
}

type ConfirmDialog struct {
	Message   string
	Action    string
	JobID     int
	ProjectID int
	focused   int // 0 = Yes, 1 = No
}

func NewConfirmDialog(message, action string, projectID, jobID int) ConfirmDialog {
	return ConfirmDialog{
		Message:   message,
		Action:    action,
		JobID:     jobID,
		ProjectID: projectID,
	}
}

func (d ConfirmDialog) Update(msg tea.Msg) (ConfirmDialog, *ConfirmResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			d.focused = 0
		case "right", "l":
			d.focused = 1
		case "enter":
			return d, &ConfirmResult{
				Confirmed: d.focused == 0,
				Action:    d.Action,
				JobID:     d.JobID,
				ProjectID: d.ProjectID,
			}
		case "y":
			return d, &ConfirmResult{Confirmed: true, Action: d.Action, JobID: d.JobID, ProjectID: d.ProjectID}
		case "n", "esc":
			return d, &ConfirmResult{Confirmed: false, Action: d.Action, JobID: d.JobID, ProjectID: d.ProjectID}
		}
	}
	return d, nil
}

func (d ConfirmDialog) View() string {
	yes := " Yes "
	no := " No "
	if d.focused == 0 {
		yes = styles.Selected.Render(yes)
	} else {
		no = styles.Selected.Render(no)
	}
	return fmt.Sprintf("\n  %s\n\n  %s  %s\n", d.Message, yes, no)
}
```

**Step 5: Install bubbletea and lipgloss**

Run: `go get github.com/charmbracelet/bubbletea github.com/charmbracelet/lipgloss github.com/charmbracelet/bubbles`

**Step 6: Verify compilation**

Run: `go build ./...`

**Step 7: Commit**

```bash
git add internal/presentation/ go.mod go.sum
git commit -m "feat: presentation — styles, breadcrumb, statusbar, confirm dialog"
```

---

### Task 8: Presentation — Views (Projects, Pipelines, Jobs, Log)

**Files:**
- Create: `internal/presentation/tui/views/projects.go`
- Create: `internal/presentation/tui/views/pipelines.go`
- Create: `internal/presentation/tui/views/jobs.go`
- Create: `internal/presentation/tui/views/log.go`

**Step 1: Create projects view**

```go
// internal/presentation/tui/views/projects.go
package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type ProjectsView struct {
	Projects []entity.Project
	Cursor   int
}

func NewProjectsView() ProjectsView {
	return ProjectsView{}
}

type ProjectSelectedMsg struct {
	Project entity.Project
}

func (v ProjectsView) Update(msg tea.Msg) (ProjectsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 {
				v.Cursor--
			}
		case "down", "j":
			if v.Cursor < len(v.Projects)-1 {
				v.Cursor++
			}
		case "enter":
			if len(v.Projects) > 0 {
				return v, func() tea.Msg {
					return ProjectSelectedMsg{Project: v.Projects[v.Cursor]}
				}
			}
		}
	}
	return v, nil
}

func (v ProjectsView) View() string {
	s := "\n"
	for i, p := range v.Projects {
		cursor := "  "
		style := styles.HelpDesc
		if i == v.Cursor {
			cursor = "▸ "
			style = styles.Selected
		}
		line := fmt.Sprintf("%s%-40s %d pipelines  %d active",
			cursor, p.PathWithNS, p.PipelineCount, p.ActiveCount)
		s += style.Render(line) + "\n"
	}
	if len(v.Projects) == 0 {
		s += styles.HelpDesc.Render("  Loading projects...") + "\n"
	}
	return s
}
```

**Step 2: Create pipelines view**

```go
// internal/presentation/tui/views/pipelines.go
package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type PipelinesView struct {
	Pipelines []entity.Pipeline
	Cursor    int
}

func NewPipelinesView() PipelinesView {
	return PipelinesView{}
}

type PipelineSelectedMsg struct {
	Pipeline entity.Pipeline
}

func (v PipelinesView) Update(msg tea.Msg) (PipelinesView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 {
				v.Cursor--
			}
		case "down", "j":
			if v.Cursor < len(v.Pipelines)-1 {
				v.Cursor++
			}
		case "enter":
			if len(v.Pipelines) > 0 {
				return v, func() tea.Msg {
					return PipelineSelectedMsg{Pipeline: v.Pipelines[v.Cursor]}
				}
			}
		}
	}
	return v, nil
}

func statusStyle(status valueobject.PipelineStatus) lipgloss.Style {
	switch status {
	case valueobject.PipelineSuccess:
		return styles.StatusSuccess
	case valueobject.PipelineFailed:
		return styles.StatusFailed
	case valueobject.PipelineRunning:
		return styles.StatusRunning
	case valueobject.PipelineManual:
		return styles.StatusManual
	default:
		return styles.StatusPending
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func (v PipelinesView) View() string {
	s := "\n"
	for i, pl := range v.Pipelines {
		cursor := "  "
		lineStyle := styles.HelpDesc
		if i == v.Cursor {
			cursor = "▸ "
			lineStyle = styles.Selected
		}
		st := statusStyle(pl.Status)
		symbol := st.Render(pl.Status.Symbol())
		status := st.Render(string(pl.Status))
		line := fmt.Sprintf("%s#%-8d %-20s %s %-12s %s",
			cursor, pl.ID, pl.Ref, symbol, status, timeAgo(pl.CreatedAt))
		_ = lineStyle
		s += line + "\n"
	}
	if len(v.Pipelines) == 0 {
		s += styles.HelpDesc.Render("  Loading pipelines...") + "\n"
	}
	return s
}
```

**Step 3: Create jobs view**

```go
// internal/presentation/tui/views/jobs.go
package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type JobsView struct {
	Jobs   []entity.Job
	Cursor int
}

func NewJobsView() JobsView {
	return JobsView{}
}

type JobSelectedMsg struct {
	Job entity.Job
}

type JobActionMsg struct {
	Action string // "play", "retry", "cancel"
	Job    entity.Job
}

func (v JobsView) Update(msg tea.Msg) (JobsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 {
				v.Cursor--
			}
		case "down", "j":
			if v.Cursor < len(v.Jobs)-1 {
				v.Cursor++
			}
		case "enter":
			if len(v.Jobs) > 0 {
				return v, func() tea.Msg {
					return JobSelectedMsg{Job: v.Jobs[v.Cursor]}
				}
			}
		case "r":
			if len(v.Jobs) > 0 {
				job := v.Jobs[v.Cursor]
				if job.Status == valueobject.JobManual {
					return v, func() tea.Msg {
						return JobActionMsg{Action: "play", Job: job}
					}
				}
				if job.Status == valueobject.JobFailed {
					return v, func() tea.Msg {
						return JobActionMsg{Action: "retry", Job: job}
					}
				}
			}
		case "c":
			if len(v.Jobs) > 0 {
				job := v.Jobs[v.Cursor]
				if job.Status.CanCancel() {
					return v, func() tea.Msg {
						return JobActionMsg{Action: "cancel", Job: job}
					}
				}
			}
		}
	}
	return v, nil
}

func jobStatusStyle(status valueobject.JobStatus) lipgloss.Style {
	switch status {
	case valueobject.JobSuccess:
		return styles.StatusSuccess
	case valueobject.JobFailed:
		return styles.StatusFailed
	case valueobject.JobRunning:
		return styles.StatusRunning
	case valueobject.JobManual:
		return styles.StatusManual
	default:
		return styles.StatusPending
	}
}

func (v JobsView) View() string {
	s := "\n"
	for i, j := range v.Jobs {
		cursor := "  "
		if i == v.Cursor {
			cursor = "▸ "
		}
		st := jobStatusStyle(j.Status)
		symbol := st.Render(j.Status.Symbol())
		status := st.Render(string(j.Status))
		dur := ""
		if j.Duration > 0 {
			dur = fmt.Sprintf("%.0fs", j.Duration)
		}
		hint := ""
		if j.Status == valueobject.JobManual {
			hint = styles.HelpKey.Render(" [r:run]")
		} else if j.Status == valueobject.JobFailed {
			hint = styles.HelpKey.Render(" [r:retry]")
		} else if j.Status.CanCancel() {
			hint = styles.HelpKey.Render(" [c:cancel]")
		}
		line := fmt.Sprintf("%s%-10s %s %-12s %-8s %s%s",
			cursor, j.Stage, symbol, j.Name, status, dur, hint)
		s += line + "\n"
	}
	if len(v.Jobs) == 0 {
		s += styles.HelpDesc.Render("  Loading jobs...") + "\n"
	}
	return s
}
```

**Step 4: Create log view**

```go
// internal/presentation/tui/views/log.go
package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type LogView struct {
	viewport viewport.Model
	content  string
	ready    bool
	jobName  string
}

func NewLogView() LogView {
	return LogView{}
}

type LogContentMsg struct {
	Content string
	JobName string
}

func (v LogView) Update(msg tea.Msg) (LogView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.viewport = viewport.New(msg.Width, msg.Height-4)
		v.viewport.SetContent(v.content)
		v.ready = true
	case LogContentMsg:
		v.content = msg.Content
		v.jobName = msg.JobName
		if v.ready {
			v.viewport.SetContent(v.content)
			v.viewport.GotoBottom()
		}
	}
	if v.ready {
		var cmd tea.Cmd
		v.viewport, cmd = v.viewport.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v LogView) View() string {
	if !v.ready {
		return styles.HelpDesc.Render("  Loading log...")
	}
	header := styles.Title.Render("Log: " + v.jobName)
	lines := []string{header, "", v.viewport.View()}
	return strings.Join(lines, "\n")
}
```

**Step 5: Verify compilation**

Run: `go build ./...`

**Step 6: Commit**

```bash
git add internal/presentation/tui/views/
git commit -m "feat: presentation — projects, pipelines, jobs, log views"
```

---

### Task 9: Presentation — App Root Model (View Router)

**Files:**
- Create: `internal/presentation/tui/app.go`

**Step 1: Create the main app model**

```go
// internal/presentation/tui/app.go
package tui

import (
	"context"
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/components"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/views"
)

type viewID int

const (
	viewProjects viewID = iota
	viewPipelines
	viewJobs
	viewLog
)

type App struct {
	cfg             *config.Config
	pipelineSvc     *service.PipelineService
	jobSvc          *service.JobService
	currentView     viewID
	breadcrumb      components.Breadcrumb
	projectsView    views.ProjectsView
	pipelinesView   views.PipelinesView
	jobsView        views.JobsView
	logView         views.LogView
	confirmDialog   *components.ConfirmDialog
	selectedProject *entity.Project
	selectedPipeline *entity.Pipeline
	width           int
	height          int
	err             error
}

func NewApp(cfg *config.Config, ps *service.PipelineService, js *service.JobService) App {
	return App{
		cfg:          cfg,
		pipelineSvc:  ps,
		jobSvc:       js,
		currentView:  viewProjects,
		breadcrumb:   components.NewBreadcrumb(),
		projectsView: views.NewProjectsView(),
		pipelinesView: views.NewPipelinesView(),
		jobsView:     views.NewJobsView(),
		logView:      views.NewLogView(),
	}
}

// Messages for async operations
type projectsLoadedMsg struct{ projects []entity.Project }
type pipelinesLoadedMsg struct{ pipelines []entity.Pipeline }
type jobsLoadedMsg struct{ jobs []entity.Job }
type logLoadedMsg struct {
	content string
	jobName string
}
type jobActionDoneMsg struct {
	job *entity.Job
	err error
}
type errMsg struct{ err error }
type tickMsg time.Time

func (a App) Init() tea.Cmd {
	return tea.Batch(a.loadProjects(), a.tick())
}

func (a App) tick() tea.Cmd {
	return tea.Tick(a.cfg.RefreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a App) loadProjects() tea.Cmd {
	return func() tea.Msg {
		projects, err := a.pipelineSvc.LoadProjects(context.Background(), a.cfg.Projects)
		if err != nil {
			return errMsg{err}
		}
		return projectsLoadedMsg{projects}
	}
}

func (a App) loadPipelines(projectID int) tea.Cmd {
	return func() tea.Msg {
		pls, err := a.pipelineSvc.ListPipelines(context.Background(), projectID)
		if err != nil {
			return errMsg{err}
		}
		return pipelinesLoadedMsg{pls}
	}
}

func (a App) loadJobs(projectID, pipelineID int) tea.Cmd {
	return func() tea.Msg {
		jobs, err := a.pipelineSvc.ListJobs(context.Background(), projectID, pipelineID)
		if err != nil {
			return errMsg{err}
		}
		return jobsLoadedMsg{jobs}
	}
}

func (a App) loadLog(projectID, jobID int, jobName string) tea.Cmd {
	return func() tea.Msg {
		rc, err := a.jobSvc.GetJobLog(context.Background(), projectID, jobID)
		if err != nil {
			return errMsg{err}
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			return errMsg{err}
		}
		return logLoadedMsg{content: string(data), jobName: jobName}
	}
}

func (a App) doJobAction(action string, projectID, jobID int) tea.Cmd {
	return func() tea.Msg {
		var job *entity.Job
		var err error
		ctx := context.Background()
		switch action {
		case "play":
			job, err = a.jobSvc.PlayJob(ctx, projectID, jobID)
		case "retry":
			job, err = a.jobSvc.RetryJob(ctx, projectID, jobID)
		case "cancel":
			job, err = a.jobSvc.CancelJob(ctx, projectID, jobID)
		}
		return jobActionDoneMsg{job, err}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle confirm dialog first
	if a.confirmDialog != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			d, result := a.confirmDialog.Update(msg)
			a.confirmDialog = &d
			if result != nil {
				a.confirmDialog = nil
				if result.Confirmed {
					return a, a.doJobAction(result.Action, result.ProjectID, result.JobID)
				}
				return a, nil
			}
			return a, nil
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.logView, _ = a.logView.Update(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "esc":
			return a, a.goBack()
		case "1":
			a.currentView = viewProjects
			a.breadcrumb.Parts = nil
			return a, a.loadProjects()
		case "2":
			if a.selectedProject != nil {
				a.currentView = viewPipelines
				a.breadcrumb.Parts = []string{a.selectedProject.PathWithNS}
				return a, a.loadPipelines(a.selectedProject.ID)
			}
		case "3":
			if a.selectedPipeline != nil {
				a.currentView = viewJobs
				return a, nil
			}
		}
		// Delegate to current view
		return a, a.delegateToView(msg)

	case projectsLoadedMsg:
		a.projectsView.Projects = msg.projects
	case pipelinesLoadedMsg:
		a.pipelinesView.Pipelines = msg.pipelines
	case jobsLoadedMsg:
		a.jobsView.Jobs = msg.jobs
	case logLoadedMsg:
		a.logView, _ = a.logView.Update(views.LogContentMsg{Content: msg.content, JobName: msg.jobName})
	case jobActionDoneMsg:
		if msg.err != nil {
			a.err = msg.err
		} else if a.selectedPipeline != nil {
			return a, a.loadJobs(a.selectedProject.ID, a.selectedPipeline.ID)
		}
	case errMsg:
		a.err = msg.err

	case tickMsg:
		return a, tea.Batch(a.refreshCurrentView(), a.tick())

	case views.ProjectSelectedMsg:
		a.selectedProject = &msg.Project
		a.currentView = viewPipelines
		a.breadcrumb.Parts = []string{msg.Project.PathWithNS}
		return a, a.loadPipelines(msg.Project.ID)

	case views.PipelineSelectedMsg:
		a.selectedPipeline = &msg.Pipeline
		a.currentView = viewJobs
		a.breadcrumb.Parts = []string{a.selectedProject.PathWithNS, fmt.Sprintf("#%d", msg.Pipeline.ID)}
		return a, a.loadJobs(a.selectedProject.ID, msg.Pipeline.ID)

	case views.JobSelectedMsg:
		a.currentView = viewLog
		a.breadcrumb.Parts = []string{a.selectedProject.PathWithNS, fmt.Sprintf("#%d", a.selectedPipeline.ID), msg.Job.Name}
		return a, a.loadLog(msg.Job.ProjectID, msg.Job.ID, msg.Job.Name)

	case views.JobActionMsg:
		actionLabel := map[string]string{"play": "Run", "retry": "Retry", "cancel": "Cancel"}
		confirm := components.NewConfirmDialog(
			fmt.Sprintf("%s job \"%s\"?", actionLabel[msg.Action], msg.Job.Name),
			msg.Action, msg.Job.ProjectID, msg.Job.ID,
		)
		a.confirmDialog = &confirm
	}

	return a, nil
}

func (a *App) delegateToView(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch a.currentView {
	case viewProjects:
		a.projectsView, cmd = a.projectsView.Update(msg)
	case viewPipelines:
		a.pipelinesView, cmd = a.pipelinesView.Update(msg)
	case viewJobs:
		a.jobsView, cmd = a.jobsView.Update(msg)
	case viewLog:
		a.logView, cmd = a.logView.Update(msg)
	}
	return cmd
}

func (a *App) goBack() tea.Cmd {
	switch a.currentView {
	case viewPipelines:
		a.currentView = viewProjects
		a.breadcrumb.Parts = nil
		return a.loadProjects()
	case viewJobs:
		a.currentView = viewPipelines
		a.breadcrumb.Parts = []string{a.selectedProject.PathWithNS}
		return nil
	case viewLog:
		a.currentView = viewJobs
		a.breadcrumb.Parts = []string{a.selectedProject.PathWithNS, fmt.Sprintf("#%d", a.selectedPipeline.ID)}
		return nil
	}
	return nil
}

func (a App) refreshCurrentView() tea.Cmd {
	switch a.currentView {
	case viewProjects:
		return a.loadProjects()
	case viewPipelines:
		if a.selectedProject != nil {
			return a.loadPipelines(a.selectedProject.ID)
		}
	case viewJobs:
		if a.selectedProject != nil && a.selectedPipeline != nil {
			return a.loadJobs(a.selectedProject.ID, a.selectedPipeline.ID)
		}
	case viewLog:
		// Log refreshes handled separately
	}
	return nil
}

func (a App) View() string {
	// Tabs
	tabs := ""
	viewNames := []string{"Projects", "Pipelines", "Jobs", "Log"}
	for i, name := range viewNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		if viewID(i) == a.currentView {
			tabs += styles.ActiveTab.Render(label)
		} else {
			tabs += styles.InactiveTab.Render(label)
		}
	}

	// Breadcrumb
	bc := a.breadcrumb.View()

	// Error
	errStr := ""
	if a.err != nil {
		errStr = styles.StatusFailed.Render(fmt.Sprintf("  Error: %v", a.err)) + "\n"
	}

	// Current view
	var content string
	switch a.currentView {
	case viewProjects:
		content = a.projectsView.View()
	case viewPipelines:
		content = a.pipelinesView.View()
	case viewJobs:
		content = a.jobsView.View()
	case viewLog:
		content = a.logView.View()
	}

	// Confirm dialog overlay
	if a.confirmDialog != nil {
		content += "\n" + a.confirmDialog.View()
	}

	// Status bar
	hints := []components.HotkeyHint{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "Enter", Desc: "select"},
		{Key: "Esc", Desc: "back"},
		{Key: "r", Desc: "run/retry"},
		{Key: "c", Desc: "cancel"},
		{Key: "q", Desc: "quit"},
	}
	sb := components.NewStatusBar(hints)

	return tabs + "\n" + bc + "\n" + errStr + content + "\n" + sb.View() + "\n"
}
```

**Step 2: Verify compilation**

Run: `go build ./...`

**Step 3: Commit**

```bash
git add internal/presentation/tui/app.go
git commit -m "feat: presentation — app root model with view routing"
```

---

### Task 10: Wire Everything in main.go

**Files:**
- Modify: `cmd/glcli/main.go`

**Step 1: Update main.go**

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	gitlabinfra "github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/gitlab"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui"
)

func main() {
	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Println("No config found. Let's set up glcli!")
		cfg, err = config.RunSetupWizard()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
	}

	client, err := gitlabinfra.NewClient(cfg.GitLabURL, cfg.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GitLab client error: %v\n", err)
		os.Exit(1)
	}

	projectRepo := gitlabinfra.NewProjectRepo(client)
	pipelineRepo := gitlabinfra.NewPipelineRepo(client)
	jobRepo := gitlabinfra.NewJobRepo(client)

	pipelineSvc := service.NewPipelineService(projectRepo, pipelineRepo)
	jobSvc := service.NewJobService(jobRepo)

	app := tui.NewApp(cfg, pipelineSvc, jobSvc)
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 2: Verify it compiles**

Run: `go build -o glcli ./cmd/glcli/`
Expected: binary `glcli` created

**Step 3: Commit**

```bash
git add cmd/glcli/main.go
git commit -m "feat: wire DI in main — config, repos, services, TUI"
```

---

### Task 11: Smoke Test

**Step 1: Create test config**

Run: `echo "gitlab_url: https://your-gitlab.com\ntoken: test\nprojects:\n  - test/project\nrefresh_interval: 5s" > ~/.glcli.yaml`

**Step 2: Run the app**

Run: `go run ./cmd/glcli/`

Verify: TUI launches, shows tabs, tries to load projects (may error on auth — that's ok, the error should display in the UI).

**Step 3: Final commit**

```bash
git add -A
git commit -m "feat: glcli v0.1.0 — GitLab TUI with DDD architecture"
```
