package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

const (
	mrFieldSource = iota
	mrFieldTarget
	mrFieldTitle
	mrFieldDescription
	mrFieldDraft
	mrFieldCount
)

type MRCreateSubmitMsg struct {
	Opts entity.CreateMROptions
}

type MRCreateCancelMsg struct{}

// MRBranchSearchMsg is sent by the view to request branch search.
type MRBranchSearchMsg struct {
	Query string
	Field int // mrFieldSource or mrFieldTarget
}

// MRBranchSearchResultMsg is returned by the app with branch results.
type MRBranchSearchResultMsg struct {
	Branches []string
	Field    int
}

type MRCreateView struct {
	fields      [mrFieldCount]string
	draft       bool
	cursor      int
	active      bool
	branches    []string  // suggestions for current branch field
	branchField int       // which field the suggestions are for
	sugCursor   int
	errMsg      string
}

func NewMRCreateView() MRCreateView {
	return MRCreateView{}
}

func (v MRCreateView) IsInputMode() bool { return v.active }

func (v *MRCreateView) Activate() {
	v.active = true
	v.cursor = mrFieldSource
	v.fields = [mrFieldCount]string{}
	v.fields[mrFieldTarget] = "main"
	v.draft = false
	v.branches = nil
	v.sugCursor = 0
	v.errMsg = ""
}

func (v MRCreateView) Update(msg tea.Msg) (MRCreateView, tea.Cmd) {
	switch msg := msg.(type) {
	case MRBranchSearchResultMsg:
		if msg.Field == v.cursor {
			v.branches = msg.Branches
			v.sugCursor = 0
		}
		return v, nil
	case tea.KeyMsg:
		return v.handleKey(msg)
	}
	return v, nil
}

func (v MRCreateView) handleKey(msg tea.KeyMsg) (MRCreateView, tea.Cmd) {
	key := msg.String()

	// Branch field with suggestions visible
	if v.hasSuggestions() {
		switch key {
		case "esc":
			v.branches = nil
			return v, nil
		case "tab", "down":
			if v.sugCursor < len(v.branches)-1 {
				v.sugCursor++
			}
			return v, nil
		case "shift+tab", "up":
			if v.sugCursor > 0 {
				v.sugCursor--
			}
			return v, nil
		case "enter":
			v.fields[v.cursor] = v.branches[v.sugCursor]
			v.branches = nil
			// Auto-advance to next field
			if v.cursor < mrFieldCount-1 {
				v.cursor++
			}
			return v, nil
		}
		// Keep typing — fall through to normal input handling
	}

	switch key {
	case "esc":
		v.active = false
		return v, func() tea.Msg { return MRCreateCancelMsg{} }
	case "tab", "down":
		v.branches = nil
		if v.cursor < mrFieldCount-1 {
			v.cursor++
		}
	case "shift+tab", "up":
		v.branches = nil
		if v.cursor > 0 {
			v.cursor--
		}
	case "enter":
		if v.cursor == mrFieldDraft {
			v.draft = !v.draft
			return v, nil
		}
		if v.cursor == mrFieldDescription {
			return v.submit()
		}
		// Move to next field
		v.branches = nil
		if v.cursor < mrFieldCount-1 {
			v.cursor++
		}
	case "ctrl+s":
		return v.submit()
	case " ":
		if v.cursor == mrFieldDraft {
			v.draft = !v.draft
			return v, nil
		}
		if v.isBranchField() {
			return v, nil // no spaces in branch names
		}
		v.fields[v.cursor] += " "
	case "backspace":
		if v.cursor != mrFieldDraft && len(v.fields[v.cursor]) > 0 {
			v.fields[v.cursor] = v.fields[v.cursor][:len(v.fields[v.cursor])-1]
			v.branches = nil
			return v, v.maybeSearchBranches()
		}
	default:
		if v.cursor != mrFieldDraft && len(key) == 1 {
			v.fields[v.cursor] += key
			v.branches = nil
			return v, v.maybeSearchBranches()
		}
	}
	return v, nil
}

func (v *MRCreateView) isBranchField() bool {
	return v.cursor == mrFieldSource || v.cursor == mrFieldTarget
}

func (v *MRCreateView) hasSuggestions() bool {
	return len(v.branches) > 0 && v.isBranchField()
}

func (v *MRCreateView) maybeSearchBranches() tea.Cmd {
	if !v.isBranchField() {
		return nil
	}
	q := v.fields[v.cursor]
	if len(q) < 1 {
		return nil
	}
	field := v.cursor
	return func() tea.Msg {
		return MRBranchSearchMsg{Query: q, Field: field}
	}
}

func (v MRCreateView) submit() (MRCreateView, tea.Cmd) {
	source := strings.TrimSpace(v.fields[mrFieldSource])
	target := strings.TrimSpace(v.fields[mrFieldTarget])
	title := strings.TrimSpace(v.fields[mrFieldTitle])

	if source == "" || target == "" || title == "" {
		v.errMsg = "Source, target and title are required"
		return v, nil
	}
	if source == target {
		v.errMsg = "Source and target branches must be different"
		return v, nil
	}

	v.errMsg = ""
	opts := entity.CreateMROptions{
		SourceBranch: source,
		TargetBranch: target,
		Title:        title,
		Description:  strings.TrimSpace(v.fields[mrFieldDescription]),
		Draft:        v.draft,
	}
	v.active = false
	return v, func() tea.Msg { return MRCreateSubmitMsg{Opts: opts} }
}

func (v MRCreateView) View() string {
	labels := [mrFieldCount]string{
		"Source Branch",
		"Target Branch",
		"Title",
		"Description",
		"Draft",
	}

	s := "\n"
	s += styles.HelpKey.Render("  Create Merge Request") + "\n\n"

	for i := 0; i < mrFieldCount; i++ {
		cursor := "  "
		if i == v.cursor {
			cursor = "▸ "
		}

		label := fmt.Sprintf("%-16s", labels[i])

		if i == mrFieldDraft {
			check := "[ ]"
			if v.draft {
				check = "[x]"
			}
			s += fmt.Sprintf("%s%s %s", cursor, styles.HelpKey.Render(label), check) + "\n"
		} else {
			value := v.fields[i]
			if i == v.cursor {
				value += "█"
			}
			s += fmt.Sprintf("%s%s %s", cursor, styles.HelpKey.Render(label), value) + "\n"
		}

		// Show branch suggestions below the active branch field
		if i == v.cursor && v.hasSuggestions() {
			for j, b := range v.branches {
				sc := "   "
				style := styles.HelpDesc
				if j == v.sugCursor {
					sc = " ▸ "
					style = styles.Selected
				}
				s += style.Render(fmt.Sprintf("  %s%s", sc, b)) + "\n"
			}
		}
	}

	if v.errMsg != "" {
		s += "\n" + styles.StatusFailed.Render("  "+v.errMsg) + "\n"
	}

	s += "\n" + styles.HelpDesc.Render("  Tab/↑↓ navigate  Enter select/next  Ctrl+S submit  Esc cancel") + "\n"
	return s
}
