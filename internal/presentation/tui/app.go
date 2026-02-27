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
	viewMRs
	viewMRDetail
	viewMRCreate
	viewCommits
)

type App struct {
	cfg              *config.Config
	pipelineSvc      *service.PipelineService
	jobSvc           *service.JobService
	mrSvc            *service.MergeRequestService
	currentView      viewID
	breadcrumb       components.Breadcrumb
	projectsView     views.ProjectsView
	pipelinesView    views.PipelinesView
	jobsView         views.JobsView
	logView          views.LogView
	mergeRequestsView views.MergeRequestsView
	mrDetailView     views.MRDetailView
	mrCreateView     views.MRCreateView
	commitsView      views.CommitsView
	confirmDialog    *components.ConfirmDialog
	selectedProject  *entity.Project
	selectedPipeline *entity.Pipeline
	selectedMR       *entity.MergeRequest
	width            int
	height           int
	err              error
	loadingStatus    string
	loading          bool
}

func NewApp(cfg *config.Config, ps *service.PipelineService, js *service.JobService, mrs *service.MergeRequestService) App {
	return App{
		cfg:               cfg,
		pipelineSvc:       ps,
		jobSvc:            js,
		mrSvc:             mrs,
		currentView:       viewPipelines,
		breadcrumb:        components.NewBreadcrumb(),
		projectsView:      views.NewProjectsView(),
		pipelinesView:     views.NewPipelinesView(),
		jobsView:          views.NewJobsView(),
		logView:           views.NewLogView(),
		mergeRequestsView: views.NewMergeRequestsView(),
		mrDetailView:      views.NewMRDetailView(),
		mrCreateView:      views.NewMRCreateView(),
		commitsView:       views.NewCommitsView(),
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
type mrsLoadedMsg struct{ mrs []entity.MergeRequest }
type mrDetailLoadedMsg struct{ mr *entity.MergeRequest }
type mrDiffsLoadedMsg struct{ diffs []entity.MRDiff }
type mrNotesLoadedMsg struct{ notes []entity.MRNote }
type commitsLoadedMsg struct{ commits []entity.Commit }
type mrCreatedMsg struct {
	mr  *entity.MergeRequest
	err error
}
type mrApprovedMsg struct{ err error }
type mrMergedMsg struct {
	mr  *entity.MergeRequest
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

func (a App) loadMRs(projectID int) tea.Cmd {
	return func() tea.Msg {
		mrs, err := a.mrSvc.ListMRs(context.Background(), projectID, "opened")
		if err != nil {
			return errMsg{err}
		}
		return mrsLoadedMsg{mrs}
	}
}

func (a App) loadAllMRs() tea.Cmd {
	return func() tea.Msg {
		// Resolve project IDs via GetByPath (fast, exact match)
		projects, err := a.pipelineSvc.LoadProjects(context.Background(), a.cfg.Projects)
		if err != nil {
			return errMsg{err}
		}
		var allMRs []entity.MergeRequest
		for _, p := range projects {
			mrs, err := a.mrSvc.ListMRs(context.Background(), p.ID, "opened")
			if err != nil {
				continue
			}
			for i := range mrs {
				mrs[i].ProjectPath = p.PathWithNS
			}
			allMRs = append(allMRs, mrs...)
		}
		return mrsLoadedMsg{allMRs}
	}
}

func (a App) loadMRDetail(projectID, mrIID int) tea.Cmd {
	return func() tea.Msg {
		mr, err := a.mrSvc.GetMR(context.Background(), projectID, mrIID)
		if err != nil {
			return errMsg{err}
		}
		return mrDetailLoadedMsg{mr}
	}
}

func (a App) loadMRDiffs(projectID, mrIID int) tea.Cmd {
	return func() tea.Msg {
		diffs, err := a.mrSvc.GetDiffs(context.Background(), projectID, mrIID)
		if err != nil {
			return errMsg{err}
		}
		return mrDiffsLoadedMsg{diffs}
	}
}

func (a App) loadMRNotes(projectID, mrIID int) tea.Cmd {
	return func() tea.Msg {
		notes, err := a.mrSvc.ListNotes(context.Background(), projectID, mrIID)
		if err != nil {
			return errMsg{err}
		}
		return mrNotesLoadedMsg{notes}
	}
}

func (a App) loadCommits(projectID int, ref string) tea.Cmd {
	return func() tea.Msg {
		commits, err := a.mrSvc.ListCommits(context.Background(), projectID, ref)
		if err != nil {
			return errMsg{err}
		}
		return commitsLoadedMsg{commits}
	}
}

func (a App) doApproveMR(projectID, mrIID int) tea.Cmd {
	return func() tea.Msg {
		err := a.mrSvc.ApproveMR(context.Background(), projectID, mrIID)
		return mrApprovedMsg{err}
	}
}

func (a App) doMergeMR(projectID, mrIID int) tea.Cmd {
	return func() tea.Msg {
		mr, err := a.mrSvc.MergeMR(context.Background(), projectID, mrIID)
		return mrMergedMsg{mr: mr, err: err}
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
					switch result.Action {
					case "approve_mr":
						return a, a.doApproveMR(result.ProjectID, result.JobID)
					case "merge_mr":
						return a, a.doMergeMR(result.ProjectID, result.JobID)
					default:
						return a, a.doJobAction(result.Action, result.ProjectID, result.JobID)
					}
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
		a.mergeRequestsView.SetHeight(msg.Height)
		a.commitsView.SetHeight(msg.Height)
		a.logView, _ = a.logView.Update(msg)
		a.mrDetailView, _ = a.mrDetailView.Update(msg)
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
			return a, a.switchToView(viewProjects)
		case "2":
			return a, a.switchToView(viewPipelines)
		case "3":
			if a.selectedPipeline != nil {
				a.currentView = viewJobs
				return a, nil
			}
		case "5":
			return a, a.switchToView(viewMRs)
		}
		// Normalize key for views
		normalizedMsg := tea.KeyMsg(tea.Key{Type: msg.Type, Runes: []rune(key)})
		if key != msg.String() {
			normalizedMsg = tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune(key)})
		}
		return a, a.delegateToView(normalizedMsg)
	case projectsLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.projectsView.Projects = msg.projects
	case allPipelinesLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.pipelinesView.Limit = a.cfg.PipelineLimit
		a.pipelinesView.SetPipelines(msg.pipelines)
	case pipelinesLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.pipelinesView.Pipelines = msg.pipelines
	case jobsLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.jobsView.Jobs = msg.jobs
		if a.jobsView.Cursor >= len(msg.jobs) {
			a.jobsView.Cursor = max(0, len(msg.jobs)-1)
		}
	case logLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
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
	case mrsLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.mergeRequestsView.SetMRs(msg.mrs)
	case mrDetailLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.mrDetailView.SetMR(msg.mr)
	case mrDiffsLoadedMsg:
		a.err = nil
		a.mrDetailView.SetDiffs(msg.diffs)
	case mrNotesLoadedMsg:
		a.err = nil
		a.mrDetailView.SetNotes(msg.notes)
	case commitsLoadedMsg:
		a.err = nil
		a.loading = false
		a.loadingStatus = ""
		a.commitsView.Commits = msg.commits
	case mrApprovedMsg:
		if msg.err != nil {
			a.err = msg.err
		} else if a.selectedMR != nil {
			return a, a.loadMRDetail(a.selectedMR.ProjectID, a.selectedMR.IID)
		}
	case mrMergedMsg:
		if msg.err != nil {
			a.err = msg.err
		} else if a.selectedMR != nil {
			return a, a.loadMRDetail(a.selectedMR.ProjectID, a.selectedMR.IID)
		}
	case views.MRCreateSubmitMsg:
		a.loading = true
		a.loadingStatus = "Creating merge request..."
		projectPath := msg.ProjectPath
		opts := msg.Opts
		return a, func() tea.Msg {
			// Resolve project ID from path
			projects, err := a.pipelineSvc.LoadProjects(context.Background(), []string{projectPath})
			if err != nil || len(projects) == 0 {
				if err == nil {
					err = fmt.Errorf("project %q not found", projectPath)
				}
				return errMsg{err}
			}
			mr, err := a.mrSvc.CreateMR(context.Background(), projects[0].ID, opts)
			return mrCreatedMsg{mr: mr, err: err}
		}
	case views.MRCreateCancelMsg:
		a.currentView = viewMRs
		a.breadcrumb.Parts = nil
	case mrCreatedMsg:
		a.loading = false
		a.loadingStatus = ""
		if msg.err != nil {
			a.err = msg.err
			a.currentView = viewMRs
			a.breadcrumb.Parts = nil
		} else {
			a.selectedMR = msg.mr
			a.currentView = viewMRDetail
			a.breadcrumb.Parts = []string{
				msg.mr.ProjectPath,
				fmt.Sprintf("!%d", msg.mr.IID),
			}
			return a, tea.Batch(
				a.loadMRDetail(msg.mr.ProjectID, msg.mr.IID),
				a.loadMRDiffs(msg.mr.ProjectID, msg.mr.IID),
				a.loadMRNotes(msg.mr.ProjectID, msg.mr.IID),
			)
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
	case views.MRBranchSearchMsg:
		projectPath := msg.ProjectPath
		field := msg.Field
		query := msg.Query
		return a, func() tea.Msg {
			projects, err := a.pipelineSvc.LoadProjects(context.Background(), []string{projectPath})
			if err != nil || len(projects) == 0 {
				return views.MRBranchSearchResultMsg{Field: field}
			}
			branches, err := a.pipelineSvc.ListBranches(context.Background(), projects[0].ID, query)
			if err != nil {
				return views.MRBranchSearchResultMsg{Field: field}
			}
			return views.MRBranchSearchResultMsg{Branches: branches, Field: field}
		}
	case views.MRBranchSearchResultMsg:
		a.mrCreateView, _ = a.mrCreateView.Update(msg)
		return a, nil
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
	case views.MRSelectedMsg:
		a.selectedMR = &msg.MR
		a.currentView = viewMRDetail
		a.breadcrumb.Parts = []string{
			msg.MR.ProjectPath,
			fmt.Sprintf("!%d", msg.MR.IID),
		}
		return a, tea.Batch(
			a.loadMRDetail(msg.MR.ProjectID, msg.MR.IID),
			a.loadMRDiffs(msg.MR.ProjectID, msg.MR.IID),
			a.loadMRNotes(msg.MR.ProjectID, msg.MR.IID),
		)
	case views.MRRefreshMsg:
		a.mrDetailView.ForceReset()
		return a, tea.Batch(
			a.loadMRDetail(msg.MR.ProjectID, msg.MR.IID),
			a.loadMRDiffs(msg.MR.ProjectID, msg.MR.IID),
			a.loadMRNotes(msg.MR.ProjectID, msg.MR.IID),
		)
	case views.MRApproveMsg:
		confirm := components.NewConfirmDialog(
			fmt.Sprintf("Approve MR !%d?", msg.MR.IID),
			"approve_mr",
			msg.MR.ProjectID,
			msg.MR.IID,
		)
		a.confirmDialog = &confirm
	case views.MRMergeMsg:
		confirm := components.NewConfirmDialog(
			fmt.Sprintf("Merge MR !%d?", msg.MR.IID),
			"merge_mr",
			msg.MR.ProjectID,
			msg.MR.IID,
		)
		a.confirmDialog = &confirm
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
	case viewMRs:
		return a.mergeRequestsView.IsInputMode()
	case viewMRCreate:
		return a.mrCreateView.IsInputMode()
	}
	return false
}

func (a *App) delegateToView(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch a.currentView {
	case viewProjects:
		// hotkey 'm' from projects to MRs
		if msg.String() == "m" {
			a.currentView = viewMRs
			a.breadcrumb.Parts = nil
			return a.loadAllMRs()
		}
		a.projectsView, cmd = a.projectsView.Update(msg)
	case viewPipelines:
		// hotkey 'c' from pipelines to commits
		if msg.String() == "c" && a.pipelinesView.Cursor < len(a.pipelinesView.Pipelines) {
			pls := a.pipelinesView.Pipelines
			if len(pls) > 0 {
				pl := pls[a.pipelinesView.Cursor]
				a.currentView = viewCommits
				a.commitsView.Ref = pl.Ref
				a.commitsView.Cursor = 0
				a.breadcrumb.Parts = []string{pl.ProjectPath, pl.Ref, "commits"}
				return a.loadCommits(pl.ProjectID, pl.Ref)
			}
		}
		a.pipelinesView, cmd = a.pipelinesView.Update(msg)
	case viewJobs:
		a.jobsView, cmd = a.jobsView.Update(msg)
	case viewLog:
		a.logView, cmd = a.logView.Update(msg)
	case viewMRs:
		if msg.String() == "n" {
			a.mrCreateView.Activate(a.cfg.Projects)
			a.currentView = viewMRCreate
			a.breadcrumb.Parts = []string{"New MR"}
			return nil
		}
		a.mergeRequestsView, cmd = a.mergeRequestsView.Update(msg)
	case viewMRCreate:
		a.mrCreateView, cmd = a.mrCreateView.Update(msg)
	case viewMRDetail:
		a.mrDetailView, cmd = a.mrDetailView.Update(msg)
	case viewCommits:
		a.commitsView, cmd = a.commitsView.Update(msg)
	}
	return cmd
}

// tabViews defines the top-level views accessible via Tab cycling.
// Sub-views (MRDetail, Commits) are reached via Enter/hotkeys, not Tab.
var tabViews = []viewID{viewProjects, viewPipelines, viewMRs}

func (a *App) tabIndex() int {
	for i, v := range tabViews {
		if v == a.currentView {
			return i
		}
	}
	// Sub-views map to their parent for tab purposes
	switch a.currentView {
	case viewJobs, viewLog, viewCommits:
		return 1 // Pipelines
	case viewMRDetail, viewMRCreate:
		return 2 // MRs
	}
	return 0
}

func (a *App) nextView() tea.Cmd {
	idx := a.tabIndex()
	idx = (idx + 1) % len(tabViews)
	return a.switchToView(tabViews[idx])
}

func (a *App) prevView() tea.Cmd {
	idx := a.tabIndex()
	idx = (idx - 1 + len(tabViews)) % len(tabViews)
	return a.switchToView(tabViews[idx])
}

func (a *App) switchToView(v viewID) tea.Cmd {
	a.currentView = v
	switch v {
	case viewProjects:
		a.breadcrumb.Parts = nil
		a.loading = true
		a.loadingStatus = fmt.Sprintf("Loading %d projects...", len(a.cfg.Projects))
		return a.loadProjects()
	case viewPipelines:
		a.breadcrumb.Parts = nil
		a.loading = true
		a.loadingStatus = fmt.Sprintf("Loading %d projects...", len(a.cfg.Projects))
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
	case viewMRs:
		a.breadcrumb.Parts = nil
		a.loading = true
		a.loadingStatus = "Loading merge requests..."
		return a.loadAllMRs()
	case viewMRDetail:
		// only reachable via enter on MR
		return nil
	case viewCommits:
		// only reachable via hotkey
		return nil
	}
	return nil
}

func (a *App) goBack() tea.Cmd {
	switch a.currentView {
	case viewPipelines:
		return a.switchToView(viewProjects)
	case viewJobs:
		a.currentView = viewPipelines
		a.breadcrumb.Parts = nil
	case viewLog:
		a.currentView = viewJobs
		a.breadcrumb.Parts = []string{
			a.selectedPipeline.ProjectPath,
			fmt.Sprintf("#%d", a.selectedPipeline.ID),
		}
	case viewMRs:
		return a.switchToView(viewProjects)
	case viewMRDetail:
		a.currentView = viewMRs
		a.breadcrumb.Parts = nil
	case viewMRCreate:
		a.currentView = viewMRs
		a.breadcrumb.Parts = nil
	case viewCommits:
		a.currentView = viewPipelines
		a.breadcrumb.Parts = nil
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
	case viewMRs:
		return a.loadAllMRs()
	case viewMRDetail:
		if a.selectedMR != nil {
			return tea.Batch(
				a.loadMRDetail(a.selectedMR.ProjectID, a.selectedMR.IID),
				a.loadMRDiffs(a.selectedMR.ProjectID, a.selectedMR.IID),
				a.loadMRNotes(a.selectedMR.ProjectID, a.selectedMR.IID),
			)
		}
	case viewCommits:
		// no auto-refresh for commits
	}
	return nil
}

func (a App) View() string {
	// Header: tabs + breadcrumb
	tabs := ""
	type tabDef struct {
		key  string
		name string
		id   viewID
	}
	tabDefs := []tabDef{
		{"1", "Projects", viewProjects},
		{"2", "Pipelines", viewPipelines},
		{"3", "Jobs", viewJobs},
		{"4", "Log", viewLog},
		{"5", "MRs", viewMRs},
	}
	for _, td := range tabDefs {
		label := fmt.Sprintf(" %s:%s ", td.key, td.name)
		if td.id == a.currentView || (a.currentView == viewMRDetail && td.id == viewMRs) ||
			(a.currentView == viewMRCreate && td.id == viewMRs) ||
			(a.currentView == viewCommits && td.id == viewPipelines) {
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
			{Key: "m", Desc: "MRs"},
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
			{Key: "c", Desc: "commits"},
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
	case viewMRs:
		hints = []components.HotkeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "Enter", Desc: "detail"},
			{Key: "n", Desc: "new MR"},
			{Key: "/", Desc: "filter"},
			{Key: "Esc", Desc: "back"},
			{Key: "q", Desc: "quit"},
		}
	case viewMRCreate:
		hints = []components.HotkeyHint{
			{Key: "Tab/↑↓", Desc: "navigate"},
			{Key: "Enter", Desc: "next/toggle"},
			{Key: "Ctrl+S", Desc: "submit"},
			{Key: "Esc", Desc: "cancel"},
		}
	case viewMRDetail:
		hints = []components.HotkeyHint{
			{Key: "↑↓", Desc: "scroll"},
			{Key: "Tab", Desc: "diff/comments"},
			{Key: "r", Desc: "refresh"},
			{Key: "a", Desc: "approve"},
			{Key: "m", Desc: "merge"},
			{Key: "Esc", Desc: "back"},
			{Key: "q", Desc: "quit"},
		}
	case viewCommits:
		hints = []components.HotkeyHint{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "Esc", Desc: "back"},
			{Key: "q", Desc: "quit"},
		}
	}
	footer := components.NewStatusBar(hints).View()

	// Content
	var content string
	switch a.currentView {
	case viewProjects:
		a.projectsView.LoadingStatus = a.loadingStatus
		content = a.projectsView.View()
	case viewPipelines:
		a.pipelinesView.LoadingStatus = a.loadingStatus
		content = a.pipelinesView.View()
	case viewJobs:
		content = a.jobsView.View()
	case viewLog:
		content = a.logView.View()
	case viewMRs:
		a.mergeRequestsView.LoadingStatus = a.loadingStatus
		content = a.mergeRequestsView.View()
	case viewMRCreate:
		content = a.mrCreateView.View()
	case viewMRDetail:
		content = a.mrDetailView.View()
	case viewCommits:
		content = a.commitsView.View()
	}
	// Fixed layout: header top, content middle, footer bottom
	// confirmDialog takes 4 lines when active, replacing footer area
	confirmLines := 0
	if a.confirmDialog != nil {
		confirmLines = 4
	}
	// headerLines=2 (tabs + breadcrumb), footerLines=1, padding=2
	contentHeight := a.height - 5 - confirmLines
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

	if a.confirmDialog != nil {
		return header + errStr + content + "\n" + a.confirmDialog.View() + "\n" + footer
	}
	return header + errStr + content + "\n" + footer
}
