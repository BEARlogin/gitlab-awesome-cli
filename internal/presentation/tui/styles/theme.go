package styles

import "github.com/charmbracelet/lipgloss"

var (
	Purple    = lipgloss.Color("99")
	Green     = lipgloss.Color("82")
	Red       = lipgloss.Color("196")
	Yellow    = lipgloss.Color("214")
	Gray      = lipgloss.Color("245")
	LightGray = lipgloss.Color("241")
	Cyan      = lipgloss.Color("87")
	White     = lipgloss.Color("255")

	Title = lipgloss.NewStyle().Bold(true).Foreground(Purple).Padding(0, 1)
	ActiveTab = lipgloss.NewStyle().Bold(true).Foreground(White).Background(Purple).Padding(0, 2)
	InactiveTab = lipgloss.NewStyle().Foreground(Gray).Padding(0, 2)
	StatusSuccess = lipgloss.NewStyle().Foreground(Green)
	StatusFailed  = lipgloss.NewStyle().Foreground(Red)
	StatusRunning = lipgloss.NewStyle().Foreground(Cyan)
	StatusManual  = lipgloss.NewStyle().Foreground(Yellow)
	StatusPending = lipgloss.NewStyle().Foreground(Gray)
	Selected = lipgloss.NewStyle().Bold(true).Foreground(White).Background(lipgloss.Color("62"))
	HelpKey = lipgloss.NewStyle().Bold(true).Foreground(Purple)
	HelpDesc = lipgloss.NewStyle().Foreground(Gray)
)
