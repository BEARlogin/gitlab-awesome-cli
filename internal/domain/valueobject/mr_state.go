package valueobject

type MRState string

const (
	MROpened MRState = "opened"
	MRMerged MRState = "merged"
	MRClosed MRState = "closed"
)

func (s MRState) Symbol() string {
	switch s {
	case MROpened:
		return "◉"
	case MRMerged:
		return "✓"
	case MRClosed:
		return "✗"
	default:
		return "?"
	}
}
