package entity

import "time"

type CreateMROptions struct {
	SourceBranch string
	TargetBranch string
	Title        string
	Description  string
	Draft        bool
}

type MergeRequest struct {
	ID           int
	IID          int
	ProjectID    int
	ProjectPath  string
	Title        string
	Description  string
	State        string
	Author       string
	SourceBranch string
	TargetBranch string
	MergeStatus  string
	Draft        bool
	WebURL       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
