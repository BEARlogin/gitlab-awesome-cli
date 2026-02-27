package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

type mrDetailTab int

const (
	mrTabDiffs mrDetailTab = iota
	mrTabComments
)

type MRDetailView struct {
	viewport    viewport.Model
	mr          *entity.MergeRequest
	diffs       []entity.MRDiff
	notes       []entity.MRNote
	tab         mrDetailTab
	ready       bool
	diffsLoaded bool
	notesLoaded bool
}

func NewMRDetailView() MRDetailView { return MRDetailView{} }

type MRApproveMsg struct{ MR entity.MergeRequest }
type MRMergeMsg struct{ MR entity.MergeRequest }
type MRRefreshMsg struct{ MR entity.MergeRequest }
type MRApprovedMsg struct{ Err error }
type MRMergedMsg struct {
	MR  *entity.MergeRequest
	Err error
}

func (v *MRDetailView) SetMR(mr *entity.MergeRequest) {
	isNewMR := v.mr == nil || v.mr.IID != mr.IID || v.mr.ProjectID != mr.ProjectID
	v.mr = mr
	if isNewMR {
		v.diffs = nil
		v.notes = nil
		v.diffsLoaded = false
		v.notesLoaded = false
		v.tab = mrTabDiffs
	}
	v.rebuildContent()
}

func (v *MRDetailView) ForceReset() {
	v.diffs = nil
	v.notes = nil
	v.diffsLoaded = false
	v.notesLoaded = false
	v.rebuildContent()
}

func (v *MRDetailView) SetDiffs(diffs []entity.MRDiff) {
	v.diffs = diffs
	v.diffsLoaded = true
	if v.tab == mrTabDiffs {
		v.rebuildContent()
	}
}

func (v *MRDetailView) SetNotes(notes []entity.MRNote) {
	v.notes = notes
	v.notesLoaded = true
	if v.tab == mrTabComments {
		v.rebuildContent()
	}
}

func (v *MRDetailView) rebuildContent() {
	if v.mr == nil {
		return
	}
	var b strings.Builder

	state := valueobject.MRState(v.mr.State)
	draft := ""
	if v.mr.Draft {
		draft = " [Draft]"
	}
	fmt.Fprintf(&b, "%s !%d: %s%s\n", state.Symbol(), v.mr.IID, v.mr.Title, draft)
	fmt.Fprintf(&b, "Author: @%s  |  %s → %s  |  %s  |  %s\n",
		v.mr.Author, v.mr.SourceBranch, v.mr.TargetBranch, v.mr.State, v.mr.MergeStatus)
	if v.mr.Description != "" {
		fmt.Fprintf(&b, "\n%s\n", v.mr.Description)
	}
	b.WriteString("\n")

	// Tab indicator
	diffTab := "  Diffs  "
	commentsTab := "  Comments  "
	if v.tab == mrTabDiffs {
		diffTab = " [Diffs] "
	} else {
		commentsTab = " [Comments] "
	}
	fmt.Fprintf(&b, "%s | %s\n", diffTab, commentsTab)
	b.WriteString(strings.Repeat("─", 60) + "\n\n")

	if v.tab == mrTabDiffs {
		if !v.diffsLoaded {
			b.WriteString("Loading diffs...\n")
		} else if len(v.diffs) == 0 {
			b.WriteString("No changes.\n")
		} else {
			for _, d := range v.diffs {
				label := d.NewPath
				if d.NewFile {
					label += " (new)"
				} else if d.DeletedFile {
					label += " (deleted)"
				} else if d.RenamedFile {
					label = fmt.Sprintf("%s → %s (renamed)", d.OldPath, d.NewPath)
				}
				b.WriteString(styles.DiffFilePath.Render(label) + "\n")
				b.WriteString(renderDiffLines(d.Diff))
				b.WriteByte('\n')
			}
		}
	} else {
		if !v.notesLoaded {
			b.WriteString("Loading comments...\n")
		} else if len(v.notes) == 0 {
			b.WriteString("No comments.\n")
		} else {
			for _, n := range v.notes {
				prefix := ""
				if n.System {
					prefix = "[system] "
				}
				fmt.Fprintf(&b, "%s@%s (%s):\n%s\n\n",
					prefix, n.Author, timeAgo(n.CreatedAt), n.Body)
			}
		}
	}

	if v.ready {
		v.viewport.SetContent(b.String())
	}
}

func (v MRDetailView) Update(msg tea.Msg) (MRDetailView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.viewport = viewport.New(msg.Width, msg.Height-4)
		v.ready = true
		v.rebuildContent()
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if v.tab == mrTabDiffs {
				v.tab = mrTabComments
			} else {
				v.tab = mrTabDiffs
			}
			v.rebuildContent()
			v.viewport.GotoTop()
			return v, nil
		case "r":
			if v.mr != nil {
				mr := *v.mr
				return v, func() tea.Msg { return MRRefreshMsg{MR: mr} }
			}
		case "a":
			if v.mr != nil && v.mr.State == "opened" {
				mr := *v.mr
				return v, func() tea.Msg { return MRApproveMsg{MR: mr} }
			}
		case "m":
			if v.mr != nil && v.mr.State == "opened" {
				mr := *v.mr
				return v, func() tea.Msg { return MRMergeMsg{MR: mr} }
			}
		}
	}
	if v.ready {
		var cmd tea.Cmd
		v.viewport, cmd = v.viewport.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v MRDetailView) View() string {
	if !v.ready {
		return styles.HelpDesc.Render("  Loading MR detail...")
	}
	title := ""
	if v.mr != nil {
		title = styles.Title.Render(fmt.Sprintf("MR !%d: %s", v.mr.IID, v.mr.Title))
	}
	return strings.Join([]string{title, "", v.viewport.View()}, "\n")
}

func renderDiffLines(diff string) string {
	lines := strings.Split(diff, "\n")
	var b strings.Builder
	for _, line := range lines {
		if len(line) == 0 {
			b.WriteByte('\n')
			continue
		}
		switch {
		case strings.HasPrefix(line, "@@"):
			b.WriteString(styles.DiffHunk.Render(line))
		case line[0] == '+':
			b.WriteString(styles.DiffAdd.Render(line))
		case line[0] == '-':
			b.WriteString(styles.DiffDel.Render(line))
		default:
			b.WriteString(line)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
