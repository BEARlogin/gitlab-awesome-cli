package entity

import "time"

type Commit struct {
	ShortID     string
	Title       string
	AuthorName  string
	AuthorEmail string
	CreatedAt   time.Time
	WebURL      string
}
