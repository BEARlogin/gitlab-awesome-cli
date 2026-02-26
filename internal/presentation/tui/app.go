package tui

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/application/service"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/infrastructure/config"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/components"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/keymap"
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
	cfg              *config.Config
	pipelineSvc      *service.PipelineService
	jobSvc           *service.JobService
	currentView      viewID
	breadcrumb       components.Breadcrumb
	projectsView     views.ProjectsView
	pipelinesView    views.PipelinesView
	jobsView         views.JobsView
	logView          views.LogView
	confirmDialog    *components.ConfirmDialog
	selectedProject  *entity.Project
	selectedPipeline *entity.Pipeline
	width            int
	height           int
	err              error
	loadingStatus    string
	loading          bool
}

func NewApp(cfg *config.Config, ps *service.PipelineService, js *service.JobService) App {
	return App{
		cfg:           cfg,
		pipelineSvc:   ps,
		jobSvc:        js,
		currentView:   viewPipelines,
		breadcrumb:    components.NewBreadcrumb(),
		projectsView:  views.NewProjectsView(),
		pipelinesView: views.NewPipelinesView(),
		jobsView:      views.NewJobsView(),
		logView:       views.NewLogView(),
	}
}

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
type loadingStatusMsg struct{ text string }
type errMsg struct{ err error }
type tickMsg time.Time

func (a App) Init() tea.Cmd {
	a.loading = true
	a.loadingStatus = fmt.Sprintf("Loading %d projects...", len(a.cfg.Projects))
	return tea.Batch(a.loadAllPipelines(), a.tick())
}

func (a App) tick() tea.Cmd {
	return tea.Tick(a.cfg.RefreshInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
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

func (a App) loadAllPipelines() tea.Cmd {
	projects := a.cfg.Projects
	limit := a.cfg.PipelineLimit
	return func() tea.Msg {
		pls, err := a.pipelineSvc.LoadAllPipelines(context.Background(), projects, limit)
		if err != nil {
			return errMsg{err}
		}
		return allPipelinesLoadedMsg{pls}
	}
}

type allPipelinesLoadedMsg struct{ pipelines []entity.Pipeline }

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
		return logLoadedMsg{
			content: string(data),
			jobName: jobName,
		}
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
		return jobActionDoneMsg{
			job: job,
			err: err,
		}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.confirmDialog != nil {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			d, result := a.confirmDialog.Update(keyMsg)
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
		a.pipelinesView.SetHeight(msg.Height)
		a.jobsView.SetHeight(msg.Height)
		a.logView, _ = a.logView.Update(msg)
	case tea.KeyMsg:
		key := keymap.Normalize(msg.String())

		// If a view is in input mode (filter, add project), delegate directly
		if a.isViewInputMode() {
			normalizedMsg := tea.KeyMsg(tea.Key{Type: msg.Type, Runes: []rune(key)})
			if key != msg.String() {
				// Don't normalize in input mode — let actual characters through
				return a, a.delegateToView(msg)
			}
			return a, a.delegateToView(normalizedMsg)
		}

		switch key {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "esc":
			return a, a.goBack()
		case "tab":
			return a, a.nextView()
		case "shift+tab":
			return a, a.prevView()
		case "1":
			a.currentView = viewProjects
			a.breadcrumb.Parts = nil
			return a, a.loadProjects()
		case "2":
			a.currentView = viewPipelines
			a.breadcrumb.Parts = nil
			return a, a.loadAllPipelines()
		case "3":
			if a.selectedPipeline != nil {
				a.currentView = viewJobs
				return a, nil
			}
		}
		// Normalize key for views
		normalizedMsg := tea.KeyMsg(tea.Key{Type: msg.Type, Runes: []rune(key)})
		if key != msg.String() {
			normalizedMsg = tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune(key)})
		}
		return a, a.delegateToView(normalizedMsg)
	case projectsLoadedMsg:
		a.err = nil
		a.projectsView.Projects = msg.projects
	case allPipelinesLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.pipelinesView.Limit = a.cfg.PipelineLimit
		a.pipelinesView.SetPipelines(msg.pipelines)
	case pipelinesLoadedMsg:
		a.err = nil
		a.pipelinesView.Pipelines = msg.pipelines
	case jobsLoadedMsg:
		a.err = nil
		a.jobsView.Jobs = msg.jobs
	case logLoadedMsg:
		a.err = nil
		a.logView, _ = a.logView.Update(views.LogContentMsg{
			Content: msg.content,
			JobName: msg.jobName,
		})
	case jobActionDoneMsg:
		if msg.err != nil {
			a.err = msg.err
		} else if a.selectedPipeline != nil {
			return a, a.loadJobs(a.selectedPipeline.ProjectID, a.selectedPipeline.ID)
		}
	case loadingStatusMsg:
		a.loadingStatus = msg.text
	case errMsg:
		a.err = msg.err
		a.loading = false
		a.loadingStatus = ""
	case tickMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, a.tick())
		if !a.loading {
			a.loading = true
			a.loadingStatus = fmt.Sprintf("Refreshing %d projects...", len(a.cfg.Projects))
			cmds = append(cmds, a.refreshCurrentView())
		}
		return a, tea.Batch(cmds...)
	case views.PipelineLimitCycleMsg:
		limits := []int{20, 50, 100, 200}
		cur := a.cfg.PipelineLimit
		next := limits[0]
		for i, l := range limits {
			if l == cur && i+1 < len(limits) {
				next = limits[i+1]
				break
			}
		}
		if cur >= limits[len(limits)-1] {
			next = limits[0]
		}
		a.cfg.PipelineLimit = next
		_ = a.cfg.Save(config.DefaultPath())
		return a, a.loadAllPipelines()
	case views.ProjectSearchMsg:
		return a, func() tea.Msg {
			results, err := a.pipelineSvc.SearchProjects(context.Background(), msg.Query)
			if err != nil {
				return errMsg{err}
			}
			return views.ProjectSearchResultMsg{Projects: results}
		}
	case views.ProjectSearchResultMsg:
		a.projectsView, _ = a.projectsView.Update(msg)
		return a, nil
	case views.ProjectAddMsg:
		a.cfg.Projects = append(a.cfg.Projects, msg.Path)
		_ = a.cfg.Save(config.DefaultPath())
		return a, tea.Batch(a.loadProjects(), a.loadAllPipelines())
	case views.ProjectDeleteMsg:
		newProjects := make([]string, 0, len(a.cfg.Projects))
		for _, p := range a.cfg.Projects {
			if p != msg.Path {
				newProjects = append(newProjects, p)
			}
		}
		a.cfg.Projects = newProjects
		_ = a.cfg.Save(config.DefaultPath())
		return a, tea.Batch(a.loadProjects(), a.loadAllPipelines())
	case views.ProjectSelectedMsg:
		a.selectedProject = &msg.Project
		a.currentView = viewPipelines
		a.breadcrumb.Parts = []string{msg.Project.PathWithNS}
		return a, a.loadAllPipelines()
	case views.PipelineSelectedMsg:
		a.selectedPipeline = &msg.Pipeline
		a.currentView = viewJobs
		a.breadcrumb.Parts = []string{
			msg.Pipeline.ProjectPath,
			fmt.Sprintf("#%d", msg.Pipeline.ID),
		}
		return a, a.loadJobs(msg.Pipeline.ProjectID, msg.Pipeline.ID)
	case views.JobSelectedMsg:
		a.currentView = viewLog
		a.breadcrumb.Parts = []string{
			a.selectedPipeline.ProjectPath,
			fmt.Sprintf("#%d", a.selectedPipeline.ID),
			msg.Job.Name,
		}
		return a, a.loadLog(msg.Job.ProjectID, msg.Job.ID, msg.Job.Name)
	case views.JobActionMsg:
		actionLabel := map[string]string{
			"play":   "Run",
			"retry":  "Retry",
			"cancel": "Cancel",
		}
		confirm := components.NewConfirmDialog(
			fmt.Sprintf("%s job \"%s\"?", actionLabel[msg.Action], msg.Job.Name),
			msg.Action,
			msg.Job.ProjectID,
			msg.Job.ID,
		)
		a.confirmDialog = &confirm
	}
	return a, nil
}

func (a *App) isViewInputMode() bool {
	switch a.currentView {
	case viewProjects:
		return a.projectsView.IsInputMode()
	case viewPipelines:
		return a.pipelinesView.IsInputMode()
	}
	return false
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

func (a *App) maxView() viewID {
	if a.selectedPipeline != nil {
		return viewLog
	}
	return viewPipelines
}

func (a *App) nextView() tea.Cmd {
	next := a.currentView + 1
	if next > a.maxView() {
		next = viewProjects
	}
	return a.switchToView(next)
}

func (a *App) prevView() tea.Cmd {
	if a.currentView == viewProjects {
		return a.switchToView(a.maxView())
	}
	return a.switchToView(a.currentView - 1)
}

func (a *App) switchToView(v viewID) tea.Cmd {
	a.currentView = v
	switch v {
	case viewProjects:
		a.breadcrumb.Parts = nil
		return a.loadProjects()
	case viewPipelines:
		a.breadcrumb.Parts = nil
		return a.loadAllPipelines()
	case viewJobs:
		if a.selectedPipeline != nil {
			a.breadcrumb.Parts = []string{
				a.selectedPipeline.ProjectPath,
				fmt.Sprintf("#%d", a.selectedPipeline.ID),
			}
			return a.loadJobs(a.selectedPipeline.ProjectID, a.selectedPipeline.ID)
		}
		a.currentView = viewPipelines
		return nil
	case viewLog:
		// only reachable if already viewing a log
		return nil
	}
	return nil
}

func (a *App) goBack() tea.Cmd {
	switch a.currentView {
	case viewPipelines:
		a.currentView = viewProjects
		a.breadcrumb.Parts = nil
		return a.loadProjects()
	case viewJobs:
		a.currentView = viewPipelines
		a.breadcrumb.Parts = nil
	case viewLog:
		a.currentView = viewJobs
		a.breadcrumb.Parts = []string{
			a.selectedPipeline.ProjectPath,
			fmt.Sprintf("#%d", a.selectedPipeline.ID),
		}
	}
	return nil
}

func (a App) refreshCurrentView() tea.Cmd {
	switch a.currentView {
	case viewProjects:
		return a.loadProjects()
	case viewPipelines:
		return a.loadAllPipelines()
	case viewJobs:
		if a.selectedPipeline != nil {
			return a.loadJobs(a.selectedPipeline.ProjectID, a.selectedPipeline.ID)
		}
	case viewLog:
		if a.selectedPipeline != nil {
			job := a.jobsView.SelectedJob()
			if job != nil {
				return a.loadLog(job.ProjectID, job.ID, job.Name)
			}
		}
	}
	return nil
}

func (a App) View() string {
	// Header: tabs + breadcrumb
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
	bc := a.breadcrumb.View()
	header := tabs + "\n" + bc + "\n"

	errStr := ""
	if a.err != nil {
		errStr = styles.StatusFailed.Render(fmt.Sprintf("  Error: %v", a.err)) + "\n"
	}

	// Footer: hotkey hints
	var hints []components.HotkeyHint
	switch a.currentView {
	case viewProjects:
		hints = []components.HotkeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "Enter", Desc: "select"},
			{Key: "a", Desc: "add"},
			{Key: "d", Desc: "delete"},
			{Key: "Tab", Desc: "next tab"},
			{Key: "q", Desc: "quit"},
		}
	case viewPipelines:
		hints = []components.HotkeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "fn↑↓", Desc: "page"},
			{Key: "Enter", Desc: "jobs"},
			{Key: "/", Desc: "filter"},
			{Key: "l", Desc: "limit"},
			{Key: "Tab", Desc: "next tab"},
			{Key: "q", Desc: "quit"},
		}
	case viewJobs:
		hints = []components.HotkeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "Enter", Desc: "log"},
			{Key: "r", Desc: "run/retry"},
			{Key: "c", Desc: "cancel"},
			{Key: "Esc", Desc: "back"},
			{Key: "q", Desc: "quit"},
		}
	case viewLog:
		hints = []components.HotkeyHint{
			{Key: "↑↓", Desc: "scroll"},
			{Key: "Esc", Desc: "back"},
			{Key: "q", Desc: "quit"},
		}
	}
	footer := components.NewStatusBar(hints).View()

	// Content
	var content string
	switch a.currentView {
	case viewProjects:
		content = a.projectsView.View()
	case viewPipelines:
		a.pipelinesView.LoadingStatus = a.loadingStatus
		content = a.pipelinesView.View()
	case viewJobs:
		content = a.jobsView.View()
	case viewLog:
		content = a.logView.View()
	}
	if a.confirmDialog != nil {
		content += "\n" + a.confirmDialog.View()
	}

	// Fixed layout: header top, content middle, footer bottom
	// headerLines=2 (tabs + breadcrumb), footerLines=1, padding=2
	contentHeight := a.height - 5
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Pad or truncate content to fill exactly contentHeight lines
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}
	content = strings.Join(contentLines, "\n")

	return header + errStr + content + "\n" + footer
}
