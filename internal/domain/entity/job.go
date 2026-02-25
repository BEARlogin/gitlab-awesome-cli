package entity

import (
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
)

type Job struct {
	ID         int
	PipelineID int
	ProjectID  int
	Name       string
	Stage      string
	Status     valueobject.JobStatus
	Duration   float64
	StartedAt  *time.Time
	FinishedAt *time.Time
	WebURL     string
}
