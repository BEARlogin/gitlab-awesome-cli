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
	offset int
	height int
}

func NewJobsView() JobsView { return JobsView{height: 20} }

func (v *JobsView) SetHeight(h int) {
	v.height = h - 6
	if v.height < 5 {
		v.height = 5
	}
}

func (v *JobsView) ensureVisible() {
	if v.Cursor < v.offset {
		v.offset = v.Cursor
	}
	if v.Cursor >= v.offset+v.height {
		v.offset = v.Cursor - v.height + 1
	}
}

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
			if v.Cursor > 0 {
				v.Cursor--
				v.ensureVisible()
			}
		case "down", "j":
			if v.Cursor < len(v.Jobs)-1 {
				v.Cursor++
				v.ensureVisible()
			}
		case "home", "g":
			v.Cursor = 0
			v.ensureVisible()
		case "end", "G":
			v.Cursor = max(0, len(v.Jobs)-1)
			v.ensureVisible()
		case "pgup", "ctrl+u":
			v.Cursor -= v.height / 2
			if v.Cursor < 0 {
				v.Cursor = 0
			}
			v.ensureVisible()
		case "pgdown", "ctrl+d":
			v.Cursor += v.height / 2
			if v.Cursor >= len(v.Jobs) {
				v.Cursor = max(0, len(v.Jobs)-1)
			}
			v.ensureVisible()
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
	total := len(v.Jobs)
	end := v.offset + v.height
	if end > total {
		end = total
	}
	visible := v.Jobs[v.offset:end]

	for i, j := range visible {
		idx := v.offset + i
		cursor := "  "
		if idx == v.Cursor {
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
	if total == 0 {
		s += styles.HelpDesc.Render("  Loading jobs...") + "\n"
	}
	if total > v.height {
		s += styles.HelpDesc.Render(fmt.Sprintf("\n  %d/%d", v.Cursor+1, total)) + "\n"
	}
	return s
}
