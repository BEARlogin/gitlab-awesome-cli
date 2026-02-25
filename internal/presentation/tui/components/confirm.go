package components

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type ConfirmResult struct {
	Confirmed bool
	Action    string
	JobID     int
	ProjectID int
}

type ConfirmDialog struct {
	Message   string
	Action    string
	JobID     int
	ProjectID int
	focused   int
}

func NewConfirmDialog(message, action string, projectID, jobID int) ConfirmDialog {
	return ConfirmDialog{Message: message, Action: action, JobID: jobID, ProjectID: projectID}
}

func (d ConfirmDialog) Update(msg tea.Msg) (ConfirmDialog, *ConfirmResult) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			d.focused = 0
		case "right", "l":
			d.focused = 1
		case "enter":
			return d, &ConfirmResult{Confirmed: d.focused == 0, Action: d.Action, JobID: d.JobID, ProjectID: d.ProjectID}
		case "y":
			return d, &ConfirmResult{Confirmed: true, Action: d.Action, JobID: d.JobID, ProjectID: d.ProjectID}
		case "n", "esc":
			return d, &ConfirmResult{Confirmed: false, Action: d.Action, JobID: d.JobID, ProjectID: d.ProjectID}
		}
	}
	return d, nil
}

func (d ConfirmDialog) View() string {
	yes := " Yes "
	no := " No "
	if d.focused == 0 {
		yes = styles.Selected.Render(yes)
	} else {
		no = styles.Selected.Render(no)
	}
	return fmt.Sprintf("\n  %s\n\n  %s  %s\n", d.Message, yes, no)
}
