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

func NewPipelinesView() PipelinesView { return PipelinesView{} }

type PipelineSelectedMsg struct{ Pipeline entity.Pipeline }

func (v PipelinesView) Update(msg tea.Msg) (PipelinesView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 { v.Cursor-- }
		case "down", "j":
			if v.Cursor < len(v.Pipelines)-1 { v.Cursor++ }
		case "enter":
			if len(v.Pipelines) > 0 {
				return v, func() tea.Msg { return PipelineSelectedMsg{Pipeline: v.Pipelines[v.Cursor]} }
			}
		}
	}
	return v, nil
}

func statusStyle(status valueobject.PipelineStatus) lipgloss.Style {
	switch status {
	case valueobject.PipelineSuccess: return styles.StatusSuccess
	case valueobject.PipelineFailed: return styles.StatusFailed
	case valueobject.PipelineRunning: return styles.StatusRunning
	case valueobject.PipelineManual: return styles.StatusManual
	default: return styles.StatusPending
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute: return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour: return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour: return fmt.Sprintf("%dh ago", int(d.Hours()))
	default: return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func (v PipelinesView) View() string {
	s := "\n"
	for i, pl := range v.Pipelines {
		cursor := "  "
		if i == v.Cursor { cursor = "▸ " }
		st := statusStyle(pl.Status)
		symbol := st.Render(pl.Status.Symbol())
		status := st.Render(string(pl.Status))
		line := fmt.Sprintf("%s#%-8d %-20s %s %-12s %s", cursor, pl.ID, pl.Ref, symbol, status, timeAgo(pl.CreatedAt))
		s += line + "\n"
	}
	if len(v.Pipelines) == 0 {
		s += styles.HelpDesc.Render("  Loading pipelines...") + "\n"
	}
	return s
}
