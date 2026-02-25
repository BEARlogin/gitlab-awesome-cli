package valueobject

type PipelineStatus string

const (
	PipelineRunning  PipelineStatus = "running"
	PipelinePending  PipelineStatus = "pending"
	PipelineSuccess  PipelineStatus = "success"
	PipelineFailed   PipelineStatus = "failed"
	PipelineCanceled PipelineStatus = "canceled"
	PipelineSkipped  PipelineStatus = "skipped"
	PipelineManual   PipelineStatus = "manual"
	PipelineCreated  PipelineStatus = "created"
)

func (s PipelineStatus) Symbol() string {
	switch s {
	case PipelineRunning:
		return "●"
	case PipelinePending, PipelineCreated:
		return "◌"
	case PipelineSuccess:
		return "✓"
	case PipelineFailed:
		return "✗"
	case PipelineCanceled:
		return "⊘"
	case PipelineSkipped:
		return "»"
	case PipelineManual:
		return "⏸"
	default:
		return "?"
	}
}

func (s PipelineStatus) IsActive() bool {
	return s == PipelineRunning || s == PipelinePending
}

type JobStatus string

const (
	JobRunning  JobStatus = "running"
	JobPending  JobStatus = "pending"
	JobSuccess  JobStatus = "success"
	JobFailed   JobStatus = "failed"
	JobCanceled JobStatus = "canceled"
	JobSkipped  JobStatus = "skipped"
	JobManual   JobStatus = "manual"
	JobCreated  JobStatus = "created"
)

func (s JobStatus) Symbol() string {
	switch s {
	case JobRunning:
		return "●"
	case JobPending, JobCreated:
		return "◌"
	case JobSuccess:
		return "✓"
	case JobFailed:
		return "✗"
	case JobCanceled:
		return "⊘"
	case JobSkipped:
		return "»"
	case JobManual:
		return "⏸"
	default:
		return "?"
	}
}

func (s JobStatus) IsActionable() bool {
	return s == JobManual || s == JobFailed
}

func (s JobStatus) CanCancel() bool {
	return s == JobRunning || s == JobPending
}
