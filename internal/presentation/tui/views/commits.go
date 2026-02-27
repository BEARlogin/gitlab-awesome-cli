package views

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type CommitsView struct {
	Commits []entity.Commit
	Cursor  int
	offset  int
	height  int
	Ref     string
}

func NewCommitsView() CommitsView { return CommitsView{height: 20} }

func (v *CommitsView) SetHeight(h int) {
	v.height = h - 6
	if v.height < 5 {
		v.height = 5
	}
}

func (v *CommitsView) ensureVisible() {
	if v.Cursor < v.offset {
		v.offset = v.Cursor
	}
	if v.Cursor >= v.offset+v.height {
		v.offset = v.Cursor - v.height + 1
	}
}

func (v CommitsView) Update(msg tea.Msg) (CommitsView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.Cursor > 0 {
				v.Cursor--
				v.ensureVisible()
			}
		case "down", "j":
			if v.Cursor < len(v.Commits)-1 {
				v.Cursor++
				v.ensureVisible()
			}
		case "home", "g":
			v.Cursor = 0
			v.ensureVisible()
		case "end", "G":
			v.Cursor = max(0, len(v.Commits)-1)
			v.ensureVisible()
		case "pgup", "ctrl+u":
			v.Cursor -= v.height / 2
			if v.Cursor < 0 {
				v.Cursor = 0
			}
			v.ensureVisible()
		case "pgdown", "ctrl+d":
			v.Cursor += v.height / 2
			if v.Cursor >= len(v.Commits) {
				v.Cursor = max(0, len(v.Commits)-1)
			}
			v.ensureVisible()
		}
	}
	return v, nil
}

func (v CommitsView) View() string {
	s := ""
	if v.Ref != "" {
		s += styles.Title.Render("  Commits: "+v.Ref) + "\n"
	}
	s += "\n"

	total := len(v.Commits)
	if v.offset >= total && total > 0 {
		v.offset = 0
	}
	end := v.offset + v.height
	if end > total {
		end = total
	}
	visible := v.Commits[v.offset:end]

	for i, c := range visible {
		idx := v.offset + i
		cursor := "  "
		if idx == v.Cursor {
			cursor = "▸ "
		}
		shortID := styles.HelpKey.Render(c.ShortID)

		title := c.Title
		if len(title) > 60 {
			title = title[:57] + "..."
		}

		line := fmt.Sprintf("%s%s  %-20s  %s  %s",
			cursor, shortID, c.AuthorName, title, timeAgo(c.CreatedAt))
		s += line + "\n"
	}

	if total == 0 {
		s += styles.HelpDesc.Render("  Loading commits...") + "\n"
	}
	if total > v.height {
		s += styles.HelpDesc.Render(fmt.Sprintf("\n  %d/%d", v.Cursor+1, total)) + "\n"
	}
	return s
}
