package models

import "time"

// WorkflowParams is the input when starting a new outreach workflow.
type WorkflowParams struct {
	CompanyName string
	Slug        string
}

// WorkflowState is the complete state of a single company's outreach workflow.
type WorkflowState struct {
	CompanyName        string
	Slug               string
	StartedAt          time.Time
	Status             string // "active", "paused", "meeting_booked"
	CurrentContact     *Contact
	OutreachAttempts   []OutreachAttempt
	AgentSuggestions   []AgentSuggestion
	WorkerRestartCount int
	LastSnapshotAt     time.Time
	MeetingBookedAt    *time.Time
	MeetingNotes       string
}

type Contact struct {
	Name     string
	Role     string
	LinkedIn string
}

type OutreachAttempt struct {
	Timestamp time.Time
	Channel   string // "email", "linkedin", "slack", "phone", "other"
	Notes     string
	Contact   string
}

type AgentSuggestion struct {
	Timestamp    time.Time
	TaskType     string // "draft_message", "suggest_contact", "next_action"
	Request      string
	Response     string
	DraftMessage string
}

// --- Signal Payloads ---

type LogOutreachSignal struct {
	Channel string
	Notes   string
}

type UpdateContactSignal struct {
	Name     string
	Role     string
	LinkedIn string
}

type RequestAgentHelpSignal struct {
	TaskType string // "draft_message" | "suggest_contact" | "next_action"
	Context  string
}

type MeetingBookedSignal struct {
	Date  time.Time
	Notes string
}
