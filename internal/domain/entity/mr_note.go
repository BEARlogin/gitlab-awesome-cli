package entity

import "time"

type MRNote struct {
	ID        int
	Author    string
	Body      string
	CreatedAt time.Time
	System    bool
}
