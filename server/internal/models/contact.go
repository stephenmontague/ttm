package models

import "time"

type Contact struct {
	Name     string
	Role     string
	LinkedIn string
	Active   bool
	AddedAt  time.Time
}

type OutreachAttempt struct {
	Timestamp time.Time
	Channel   string // "email", "linkedin", "slack", "phone", "other"
	Notes     string
	Contact   string
}
