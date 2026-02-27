package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/entity"
	"github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"
)

const (
	mrFieldProject = iota
	mrFieldSource
	mrFieldTarget
	mrFieldTitle
	mrFieldDescription
	mrFieldDraft
	mrFieldCount
)

type MRCreateSubmitMsg struct {
	ProjectPath string
	Opts        entity.CreateMROptions
}

type MRCreateCancelMsg struct{}

// MRBranchSearchMsg is sent by the view to request branch search.
type MRBranchSearchMsg struct {
	ProjectPath string
	Query       string
	Field       int // mrFieldSource or mrFieldTarget
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
	projects    []string  // configured project paths
	projSugs    []string  // filtered project suggestions
	branches    []string  // suggestions for current branch field
	sugCursor   int
	errMsg      string
}

func NewMRCreateView() MRCreateView {
	return MRCreateView{}
}

func (v MRCreateView) IsInputMode() bool { return v.active }

func (v *MRCreateView) Activate(projects []string) {
	v.active = true
	v.cursor = mrFieldProject
	v.fields = [mrFieldCount]string{}
	v.fields[mrFieldTarget] = "main"
	v.draft = false
	v.projects = projects
	v.projSugs = projects // show all initially
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

	// Suggestion list active (project or branch)
	if v.hasSuggestions() {
		sugs := v.currentSuggestions()
		switch key {
		case "esc":
			v.clearSuggestions()
			return v, nil
		case "tab", "down":
			if v.sugCursor < len(sugs)-1 {
				v.sugCursor++
			}
			return v, nil
		case "shift+tab", "up":
			if v.sugCursor > 0 {
				v.sugCursor--
			}
			return v, nil
		case "enter":
			if v.sugCursor < len(sugs) {
				v.fields[v.cursor] = sugs[v.sugCursor]
				v.clearSuggestions()
				if v.cursor < mrFieldCount-1 {
					v.cursor++
				}
			}
			return v, nil
		}
		// fall through for typing
	}

	switch key {
	case "esc":
		v.active = false
		return v, func() tea.Msg { return MRCreateCancelMsg{} }
	case "tab", "down":
		v.clearSuggestions()
		if v.cursor < mrFieldCount-1 {
			v.cursor++
		}
	case "shift+tab", "up":
		v.clearSuggestions()
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
		v.clearSuggestions()
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
		if v.cursor == mrFieldProject || v.isBranchField() {
			return v, nil // no spaces in project/branch names
		}
		v.fields[v.cursor] += " "
	case "backspace":
		if v.cursor != mrFieldDraft && len(v.fields[v.cursor]) > 0 {
			v.fields[v.cursor] = v.fields[v.cursor][:len(v.fields[v.cursor])-1]
			v.clearSuggestions()
			return v, v.onFieldChanged()
		}
	default:
		if v.cursor != mrFieldDraft && len(key) == 1 {
			v.fields[v.cursor] += key
			v.clearSuggestions()
			return v, v.onFieldChanged()
		}
	}
	return v, nil
}

func (v *MRCreateView) isBranchField() bool {
	return v.cursor == mrFieldSource || v.cursor == mrFieldTarget
}

func (v *MRCreateView) hasSuggestions() bool {
	return len(v.currentSuggestions()) > 0
}

func (v *MRCreateView) currentSuggestions() []string {
	if v.cursor == mrFieldProject {
		return v.projSugs
	}
	if v.isBranchField() {
		return v.branches
	}
	return nil
}

func (v *MRCreateView) clearSuggestions() {
	v.branches = nil
	v.projSugs = nil
	v.sugCursor = 0
}

func (v *MRCreateView) onFieldChanged() tea.Cmd {
	if v.cursor == mrFieldProject {
		v.filterProjects()
		return nil
	}
	if v.isBranchField() {
		return v.searchBranches()
	}
	return nil
}

func (v *MRCreateView) filterProjects() {
	q := strings.ToLower(v.fields[mrFieldProject])
	if q == "" {
		v.projSugs = v.projects
		v.sugCursor = 0
		return
	}
	v.projSugs = nil
	for _, p := range v.projects {
		if strings.Contains(strings.ToLower(p), q) {
			v.projSugs = append(v.projSugs, p)
		}
	}
	v.sugCursor = 0
}

func (v *MRCreateView) searchBranches() tea.Cmd {
	q := v.fields[v.cursor]
	if len(q) < 1 {
		return nil
	}
	proj := v.fields[mrFieldProject]
	if proj == "" {
		return nil
	}
	field := v.cursor
	return func() tea.Msg {
		return MRBranchSearchMsg{ProjectPath: proj, Query: q, Field: field}
	}
}

func (v MRCreateView) submit() (MRCreateView, tea.Cmd) {
	project := strings.TrimSpace(v.fields[mrFieldProject])
	source := strings.TrimSpace(v.fields[mrFieldSource])
	target := strings.TrimSpace(v.fields[mrFieldTarget])
	title := strings.TrimSpace(v.fields[mrFieldTitle])

	if project == "" {
		v.errMsg = "Project is required"
		return v, nil
	}
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
	return v, func() tea.Msg { return MRCreateSubmitMsg{ProjectPath: project, Opts: opts} }
}

func (v MRCreateView) View() string {
	labels := [mrFieldCount]string{
		"Project",
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

		// Show suggestions below the active field
		if i == v.cursor {
			sugs := v.currentSuggestions()
			for j, sg := range sugs {
				sc := "   "
				style := styles.HelpDesc
				if j == v.sugCursor {
					sc = " ▸ "
					style = styles.Selected
				}
				s += style.Render(fmt.Sprintf("  %s%s", sc, sg)) + "\n"
			}
		}
	}

	if v.errMsg != "" {
		s += "\n" + styles.StatusFailed.Render("  "+v.errMsg) + "\n"
	}

	s += "\n" + styles.HelpDesc.Render("  Tab/↑↓ navigate  Enter select/next  Ctrl+S submit  Esc cancel") + "\n"
	return s
}
