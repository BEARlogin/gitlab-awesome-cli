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

func NewProjectsView() ProjectsView { return ProjectsView{} }

type ProjectSelectedMsg struct{ Project entity.Project }

func (v ProjectsView) Update(msg tea.Msg) (ProjectsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 { v.Cursor-- }
		case "down", "j":
			if v.Cursor < len(v.Projects)-1 { v.Cursor++ }
		case "enter":
			if len(v.Projects) > 0 {
				return v, func() tea.Msg { return ProjectSelectedMsg{Project: v.Projects[v.Cursor]} }
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
		line := fmt.Sprintf("%s%-40s %d pipelines  %d active", cursor, p.PathWithNS, p.PipelineCount, p.ActiveCount)
		s += style.Render(line) + "\n"
	}
	if len(v.Projects) == 0 {
		s += styles.HelpDesc.Render("  Loading projects...") + "\n"
	}
	return s
}
