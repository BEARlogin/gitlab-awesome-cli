package views

import (
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type LogView struct {
	viewport viewport.Model
	content  string
	ready    bool
	jobName  string
}

func NewLogView() LogView { return LogView{} }

type LogContentMsg struct {
	Content string
	JobName string
}

func (v LogView) Update(msg tea.Msg) (LogView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.viewport = viewport.New(msg.Width, msg.Height-4)
		v.viewport.SetContent(v.content)
		v.ready = true
	case LogContentMsg:
		v.content = msg.Content
		v.jobName = msg.JobName
		if v.ready {
			v.viewport.SetContent(v.content)
			v.viewport.GotoBottom()
		}
	}
	if v.ready {
		var cmd tea.Cmd
		v.viewport, cmd = v.viewport.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v LogView) View() string {
	if !v.ready { return styles.HelpDesc.Render("  Loading log...") }
	header := styles.Title.Render("Log: " + v.jobName)
	return strings.Join([]string{header, "", v.viewport.View()}, "\n")
}
