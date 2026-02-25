package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type ProjectsView struct {
	Projects    []entity.Project
	Cursor      int
	adding      bool
	input       string
	suggestions []entity.Project
	sugCursor   int
}

func NewProjectsView() ProjectsView { return ProjectsView{} }

type ProjectSelectedMsg struct{ Project entity.Project }
type ProjectAddMsg struct{ Path string }
type ProjectDeleteMsg struct{ Path string }

// ProjectSearchMsg is sent by the view to request a search.
// The app handles it and returns ProjectSearchResultMsg.
type ProjectSearchMsg struct{ Query string }
type ProjectSearchResultMsg struct{ Projects []entity.Project }

func (v ProjectsView) Update(msg tea.Msg) (ProjectsView, tea.Cmd) {
	switch msg := msg.(type) {
	case ProjectSearchResultMsg:
		v.suggestions = msg.Projects
		v.sugCursor = 0
		return v, nil
	case tea.KeyMsg:
		if v.adding {
			return v.updateAdding(msg)
		}
		return v.updateNormal(msg)
	}
	return v, nil
}

func (v ProjectsView) updateAdding(msg tea.KeyMsg) (ProjectsView, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(v.suggestions) > 0 && v.sugCursor < len(v.suggestions) {
			path := v.suggestions[v.sugCursor].PathWithNS
			v.adding = false
			v.input = ""
			v.suggestions = nil
			return v, func() tea.Msg { return ProjectAddMsg{Path: path} }
		}
		if v.input != "" {
			path := v.input
			v.adding = false
			v.input = ""
			v.suggestions = nil
			return v, func() tea.Msg { return ProjectAddMsg{Path: path} }
		}
	case "esc":
		v.adding = false
		v.input = ""
		v.suggestions = nil
	case "tab", "down":
		if len(v.suggestions) > 0 && v.sugCursor < len(v.suggestions)-1 {
			v.sugCursor++
		}
	case "shift+tab", "up":
		if v.sugCursor > 0 {
			v.sugCursor--
		}
	case "backspace":
		if len(v.input) > 0 {
			v.input = v.input[:len(v.input)-1]
			if len(v.input) >= 2 {
				q := v.input
				return v, func() tea.Msg { return ProjectSearchMsg{Query: q} }
			}
			v.suggestions = nil
		}
	default:
		if len(msg.String()) == 1 {
			v.input += msg.String()
			if len(v.input) >= 2 {
				q := v.input
				return v, func() tea.Msg { return ProjectSearchMsg{Query: q} }
			}
		}
	}
	return v, nil
}

func (v ProjectsView) updateNormal(msg tea.KeyMsg) (ProjectsView, tea.Cmd) {
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
			return v, func() tea.Msg { return ProjectSelectedMsg{Project: v.Projects[v.Cursor]} }
		}
	case "a":
		v.adding = true
		v.input = ""
		v.suggestions = nil
		v.sugCursor = 0
	case "d":
		if len(v.Projects) > 0 {
			p := v.Projects[v.Cursor]
			return v, func() tea.Msg { return ProjectDeleteMsg{Path: p.PathWithNS} }
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
		s += styles.HelpDesc.Render("  No projects configured") + "\n"
	}
	s += "\n"
	if v.adding {
		s += styles.HelpKey.Render("  Add project: ") + v.input + "█\n"
		if len(v.suggestions) > 0 {
			s += "\n"
			for i, p := range v.suggestions {
				cursor := "   "
				style := styles.HelpDesc
				if i == v.sugCursor {
					cursor = " ▸ "
					style = styles.Selected
				}
				s += style.Render(fmt.Sprintf("%s%s", cursor, p.PathWithNS)) + "\n"
			}
		} else if len(v.input) >= 2 {
			s += styles.HelpDesc.Render("   Searching...") + "\n"
		}
		s += "\n" + styles.HelpDesc.Render("  ↑↓ select  Enter confirm  Esc cancel") + "\n"
	} else {
		s += styles.HelpDesc.Render("  [a] add project  [d] delete project") + "\n"
	}
	return s
}
