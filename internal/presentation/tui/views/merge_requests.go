package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

type MergeRequestsView struct {
	MRs           []entity.MergeRequest
	filtered      []entity.MergeRequest
	Cursor        int
	offset        int
	height        int
	Filter        string
	filtering     bool
	loaded        bool
	LoadingStatus string
}

func NewMergeRequestsView() MergeRequestsView { return MergeRequestsView{height: 20} }

func (v MergeRequestsView) IsInputMode() bool { return v.filtering }

type MRSelectedMsg struct{ MR entity.MergeRequest }

func (v *MergeRequestsView) SetHeight(h int) {
	v.height = h - 6
	if v.height < 5 {
		v.height = 5
	}
}

func (v *MergeRequestsView) applyFilter() {
	if v.Filter == "" {
		v.filtered = v.MRs
		return
	}
	f := strings.ToLower(v.Filter)
	v.filtered = nil
	for _, mr := range v.MRs {
		if strings.Contains(strings.ToLower(mr.Title), f) ||
			strings.Contains(strings.ToLower(mr.Author), f) ||
			strings.Contains(strings.ToLower(mr.SourceBranch), f) ||
			strings.Contains(strings.ToLower(mr.ProjectPath), f) {
			v.filtered = append(v.filtered, mr)
		}
	}
	if v.Cursor >= len(v.filtered) {
		v.Cursor = max(0, len(v.filtered)-1)
	}
	v.ensureVisible()
}

func (v *MergeRequestsView) ensureVisible() {
	if v.Cursor < v.offset {
		v.offset = v.Cursor
	}
	if v.Cursor >= v.offset+v.height {
		v.offset = v.Cursor - v.height + 1
	}
}

func (v *MergeRequestsView) VisibleMRs() []entity.MergeRequest {
	return v.filtered
}

func (v *MergeRequestsView) Reset() {
	v.MRs = nil
	v.filtered = nil
	v.loaded = false
	v.Cursor = 0
	v.offset = 0
}

func (v *MergeRequestsView) SetMRs(mrs []entity.MergeRequest) {
	v.MRs = mrs
	v.loaded = true
	v.applyFilter()
}

func (v MergeRequestsView) Update(msg tea.Msg) (MergeRequestsView, tea.Cmd) {
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
			if len(v.filtered) > 0 && v.Cursor < len(v.filtered) {
				return v, func() tea.Msg { return MRSelectedMsg{MR: v.filtered[v.Cursor]} }
			}
		case "/":
			v.filtering = true
			v.Filter = ""
			v.applyFilter()
		}
	}
	return v, nil
}

func mrStateStyle(state string) lipgloss.Style {
	switch valueobject.MRState(state) {
	case valueobject.MRMerged:
		return styles.StatusSuccess
	case valueobject.MRClosed:
		return styles.StatusFailed
	default:
		return styles.StatusRunning
	}
}

func (v MergeRequestsView) View() string {
	s := ""
	if v.filtering {
		s += styles.HelpKey.Render("  Filter: ") + v.Filter + "█\n"
	} else if v.Filter != "" {
		s += styles.HelpKey.Render("  Filter: ") + styles.HelpDesc.Render(v.Filter) + "\n"
	}
	s += "\n"

	total := len(v.filtered)
	if v.offset >= total && total > 0 {
		v.offset = 0
	}
	end := v.offset + v.height
	if end > total {
		end = total
	}
	visible := v.filtered[v.offset:end]

	for i, mr := range visible {
		idx := v.offset + i
		cursor := "  "
		if idx == v.Cursor {
			cursor = "▸ "
		}
		st := mrStateStyle(mr.State)
		state := valueobject.MRState(mr.State)
		symbol := st.Render(state.Symbol())
		stateStr := st.Render(mr.State)

		proj := mr.ProjectPath
		if pidx := strings.LastIndex(proj, "/"); pidx >= 0 {
			proj = proj[pidx+1:]
		}

		title := mr.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		draft := ""
		if mr.Draft {
			draft = styles.HelpDesc.Render("[Draft] ")
		}

		line := fmt.Sprintf("%s%-16s !%-6d %-20s %s %s %-10s @%-12s %s",
			cursor, proj, mr.IID, mr.SourceBranch, symbol, draft, stateStr, mr.Author, title)
		s += line + "\n"
	}

	if total == 0 {
		if !v.loaded {
			if v.LoadingStatus != "" {
				s += styles.HelpDesc.Render("  "+v.LoadingStatus) + "\n"
			} else {
				s += styles.HelpDesc.Render("  Loading merge requests...") + "\n"
			}
		} else if v.Filter != "" {
			s += styles.HelpDesc.Render("  No merge requests match filter") + "\n"
		} else {
			s += styles.HelpDesc.Render("  No open merge requests") + "\n"
		}
	}

	if total > 0 {
		s += "\n" + styles.HelpDesc.Render(fmt.Sprintf("  %d/%d", v.Cursor+1, total)) + "\n"
	}

	return s
}
