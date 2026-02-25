package entity

type Project struct {
	ID            int
	Name          string
	PathWithNS    string
	WebURL        string
	PipelineCount int
	ActiveCount   int
}
