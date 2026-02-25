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
	adding   bool
	input    string
}

func NewProjectsView() ProjectsView { return ProjectsView{} }

type ProjectSelectedMsg struct{ Project entity.Project }
type ProjectAddMsg struct{ Path string }
type ProjectDeleteMsg struct{ Path string }

func (v ProjectsView) Update(msg tea.Msg) (ProjectsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if v.adding {
			switch msg.String() {
			case "enter":
				path := v.input
				v.adding = false
				v.input = ""
				if path != "" {
					return v, func() tea.Msg { return ProjectAddMsg{Path: path} }
				}
			case "esc":
				v.adding = false
				v.input = ""
			case "backspace":
				if len(v.input) > 0 {
					v.input = v.input[:len(v.input)-1]
				}
			default:
				if len(msg.String()) == 1 {
					v.input += msg.String()
				}
			}
			return v, nil
		}
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
		case "d":
			if len(v.Projects) > 0 {
				p := v.Projects[v.Cursor]
				return v, func() tea.Msg { return ProjectDeleteMsg{Path: p.PathWithNS} }
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
		s += styles.HelpDesc.Render("  No projects configured") + "\n"
	}
	s += "\n"
	if v.adding {
		s += styles.HelpKey.Render("  Add project: ") + v.input + "█\n"
	} else {
		s += styles.HelpDesc.Render("  [a] add project  [d] delete project") + "\n"
	}
	return s
}
