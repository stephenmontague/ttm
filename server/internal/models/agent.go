package models

import "time"

type AgentSuggestion struct {
	Timestamp    time.Time
	TaskType     string // "draft_message", "suggest_contact", "next_action"
	Request      string
	Response     string
	DraftMessage string
}
