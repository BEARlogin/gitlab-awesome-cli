package entity

type MRDiff struct {
	OldPath     string
	NewPath     string
	Diff        string
	NewFile     bool
	DeletedFile bool
	RenamedFile bool
}
