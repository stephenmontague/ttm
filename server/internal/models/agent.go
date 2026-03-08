package models

import "time"

type AgentSuggestion struct {
	Timestamp    time.Time
	TaskType     string // "draft_message", "suggest_contact", "next_action"
	ContactName  string // optional — which contact this suggestion is for
	Request      string
	Response     string
	DraftMessage string
}
