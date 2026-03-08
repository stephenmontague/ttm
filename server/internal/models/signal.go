package models

import "time"

type LogOutreachSignal struct {
	Channel     string
	Notes       string
	ContactName string // Explicit contact selection.
}

// UpdateContactSignal is deprecated; kept for backward compat with in-flight workflows.
type UpdateContactSignal struct {
	Name     string
	Role     string
	LinkedIn string
}

type AddContactSignal struct {
	Name     string
	Role     string
	LinkedIn string
}

type RemoveContactSignal struct {
	Name string
}

type RequestAgentHelpSignal struct {
	TaskType    string // "draft_message" | "suggest_contact" | "next_action"
	Context     string
	ContactName string // optional — target contact for this request
}

type MeetingBookedSignal struct {
	Date  time.Time
	Notes string
}
