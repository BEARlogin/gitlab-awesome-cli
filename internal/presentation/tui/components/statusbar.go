package components

import "github.com/bearlogin/gitlab-awesome-cli/internal/presentation/tui/styles"

type HotkeyHint struct {
	Key  string
	Desc string
}

type StatusBar struct {
	Hints []HotkeyHint
}

func NewStatusBar(hints []HotkeyHint) StatusBar {
	return StatusBar{Hints: hints}
}

func (s StatusBar) View() string {
	var result string
	for i, h := range s.Hints {
		if i > 0 { result += "  " }
		result += styles.HelpKey.Render(h.Key) + " " + styles.HelpDesc.Render(h.Desc)
	}
	return result
}
