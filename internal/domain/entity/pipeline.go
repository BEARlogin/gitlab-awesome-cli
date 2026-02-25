package entity

import (
	"time"

	"github.com/bearlogin/gitlab-awesome-cli/internal/domain/valueobject"
)

type Pipeline struct {
	ID          int
	ProjectID   int
	ProjectPath string
	Ref         string
	Status      valueobject.PipelineStatus
	CreatedAt   time.Time
	Duration    int
	JobCount    int
}
