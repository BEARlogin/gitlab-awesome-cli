package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type PipelinesView struct {
	Pipelines []entity.Pipeline
	filtered  []entity.Pipeline
	Cursor    int
	offset    int
	height    int // visible rows
	Limit     int
	Filter    string
	filtering bool
}

func NewPipelinesView() PipelinesView { return PipelinesView{height: 20} }

type PipelineSelectedMsg struct{ Pipeline entity.Pipeline }
type PipelineLimitCycleMsg struct{}

func (v *PipelinesView) SetHeight(h int) {
	// subtract header lines (filter + padding + statusbar)
	v.height = h - 6
	if v.height < 5 {
		v.height = 5
	}
}

func (v *PipelinesView) applyFilter() {
	if v.Filter == "" {
		v.filtered = v.Pipelines
		return
	}
	f := strings.ToLower(v.Filter)
	v.filtered = nil
	for _, pl := range v.Pipelines {
		if strings.Contains(strings.ToLower(pl.ProjectPath), f) ||
			strings.Contains(strings.ToLower(pl.Ref), f) ||
			strings.Contains(strings.ToLower(string(pl.Status)), f) {
			v.filtered = append(v.filtered, pl)
		}
	}
	if v.Cursor >= len(v.filtered) {
		v.Cursor = max(0, len(v.filtered)-1)
	}
	v.ensureVisible()
}

func (v *PipelinesView) ensureVisible() {
	if v.Cursor < v.offset {
		v.offset = v.Cursor
	}
	if v.Cursor >= v.offset+v.height {
		v.offset = v.Cursor - v.height + 1
	}
}

func (v PipelinesView) Update(msg tea.Msg) (PipelinesView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if v.filtering {
			switch msg.String() {
			case "enter", "esc":
				v.filtering = false
			case "backspace":
				if len(v.Filter) > 0 {
					v.Filter = v.Filter[:len(v.Filter)-1]
					v.applyFilter()
				}
			default:
				if len(msg.String()) == 1 {
					v.Filter += msg.String()
					v.applyFilter()
				}
			}
			return v, nil
		}
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 {
				v.Cursor--
				v.ensureVisible()
			}
		case "down", "j":
			if v.Cursor < len(v.filtered)-1 {
				v.Cursor++
				v.ensureVisible()
			}
		case "home", "g":
			v.Cursor = 0
			v.ensureVisible()
		case "end", "G":
			v.Cursor = max(0, len(v.filtered)-1)
			v.ensureVisible()
		case "pgup", "ctrl+u":
			v.Cursor -= v.height / 2
			if v.Cursor < 0 {
				v.Cursor = 0
			}
			v.ensureVisible()
		case "pgdown", "ctrl+d":
			v.Cursor += v.height / 2
			if v.Cursor >= len(v.filtered) {
				v.Cursor = max(0, len(v.filtered)-1)
			}
			v.ensureVisible()
		case "enter":
			if len(v.filtered) > 0 {
				return v, func() tea.Msg { return PipelineSelectedMsg{Pipeline: v.filtered[v.Cursor]} }
			}
		case "/":
			v.filtering = true
			v.Filter = ""
			v.applyFilter()
		case "l":
			return v, func() tea.Msg { return PipelineLimitCycleMsg{} }
		}
	}
	return v, nil
}

func (v *PipelinesView) SetPipelines(pls []entity.Pipeline) {
	v.Pipelines = pls
	v.applyFilter()
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
	s := ""
	if v.filtering {
		s += styles.HelpKey.Render("  Filter: ") + v.Filter + "█\n"
	} else if v.Filter != "" {
		s += styles.HelpKey.Render("  Filter: ") + styles.HelpDesc.Render(v.Filter) + "\n"
	}
	s += "\n"

	total := len(v.filtered)
	end := v.offset + v.height
	if end > total {
		end = total
	}
	visible := v.filtered[v.offset:end]

	for i, pl := range visible {
		idx := v.offset + i
		cursor := "  "
		if idx == v.Cursor {
			cursor = "▸ "
		}
		st := statusStyle(pl.Status)
		symbol := st.Render(pl.Status.Symbol())
		status := st.Render(string(pl.Status))

		proj := pl.ProjectPath
		if pidx := strings.LastIndex(proj, "/"); pidx >= 0 {
			proj = proj[pidx+1:]
		}

		line := fmt.Sprintf("%s%-16s #%-8d %-16s %s %-12s %s",
			cursor, proj, pl.ID, pl.Ref, symbol, status, timeAgo(pl.CreatedAt))
		s += line + "\n"
	}

	if total == 0 && len(v.Pipelines) == 0 {
		s += styles.HelpDesc.Render("  Loading pipelines...") + "\n"
	}
	if total == 0 && len(v.Pipelines) > 0 {
		s += styles.HelpDesc.Render("  No pipelines match filter") + "\n"
	}

	// scroll indicator + limit
	if total > 0 {
		info := fmt.Sprintf("  %d/%d", v.Cursor+1, total)
		if v.Limit > 0 {
			info += fmt.Sprintf("  limit:%d", v.Limit)
		}
		s += "\n" + styles.HelpDesc.Render(info) + "\n"
	}

	return s
}
