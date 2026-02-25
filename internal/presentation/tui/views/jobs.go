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

func NewJobsView() JobsView { return JobsView{} }

type JobSelectedMsg struct{ Job entity.Job }
type JobActionMsg struct {
	Action string
	Job    entity.Job
}

func (v JobsView) Update(msg tea.Msg) (JobsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 { v.Cursor-- }
		case "down", "j":
			if v.Cursor < len(v.Jobs)-1 { v.Cursor++ }
		case "enter":
			if len(v.Jobs) > 0 {
				return v, func() tea.Msg { return JobSelectedMsg{Job: v.Jobs[v.Cursor]} }
			}
		case "r":
			if len(v.Jobs) > 0 {
				job := v.Jobs[v.Cursor]
				if job.Status == valueobject.JobManual {
					return v, func() tea.Msg { return JobActionMsg{Action: "play", Job: job} }
				}
				if job.Status == valueobject.JobFailed {
					return v, func() tea.Msg { return JobActionMsg{Action: "retry", Job: job} }
				}
			}
		case "c":
			if len(v.Jobs) > 0 {
				job := v.Jobs[v.Cursor]
				if job.Status.CanCancel() {
					return v, func() tea.Msg { return JobActionMsg{Action: "cancel", Job: job} }
				}
			}
		}
	}
	return v, nil
}

func jobStatusStyle(status valueobject.JobStatus) lipgloss.Style {
	switch status {
	case valueobject.JobSuccess: return styles.StatusSuccess
	case valueobject.JobFailed: return styles.StatusFailed
	case valueobject.JobRunning: return styles.StatusRunning
	case valueobject.JobManual: return styles.StatusManual
	default: return styles.StatusPending
	}
}

func (v JobsView) View() string {
	s := "\n"
	for i, j := range v.Jobs {
		cursor := "  "
		if i == v.Cursor { cursor = "▸ " }
		st := jobStatusStyle(j.Status)
		symbol := st.Render(j.Status.Symbol())
		status := st.Render(string(j.Status))
		dur := ""
		if j.Duration > 0 { dur = fmt.Sprintf("%.0fs", j.Duration) }
		hint := ""
		if j.Status == valueobject.JobManual { hint = styles.HelpKey.Render(" [r:run]") }
		if j.Status == valueobject.JobFailed { hint = styles.HelpKey.Render(" [r:retry]") }
		if j.Status.CanCancel() { hint = styles.HelpKey.Render(" [c:cancel]") }
		line := fmt.Sprintf("%s%-10s %s %-12s %-8s %s%s", cursor, j.Stage, symbol, j.Name, status, dur, hint)
		s += line + "\n"
	}
	if len(v.Jobs) == 0 {
		s += styles.HelpDesc.Render("  Loading jobs...") + "\n"
	}
	return s
}
