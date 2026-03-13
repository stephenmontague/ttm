package models

import "time"

// WorkflowParams is the input when starting a new outreach workflow.
type WorkflowParams struct {
	CompanyName      string
	Slug             string
	StartedAt        time.Time          // Preserved across continue-as-new; zero on first run.
	Contacts         []Contact          // Preserved across continue-as-new.
	OutreachAttempts []OutreachAttempt   // Preserved across continue-as-new.
	AgentSuggestions []AgentSuggestion   // Preserved across continue-as-new.
	WorkerRestartCount int              // Preserved across continue-as-new.
}

// NewWorkflowState creates an initial workflow state from params.
func NewWorkflowState(params WorkflowParams, startedAt time.Time) *WorkflowState {
	contacts := params.Contacts
	if contacts == nil {
		contacts = []Contact{}
	}
	outreachAttempts := params.OutreachAttempts
	if outreachAttempts == nil {
		outreachAttempts = []OutreachAttempt{}
	}
	agentSuggestions := params.AgentSuggestions
	if agentSuggestions == nil {
		agentSuggestions = []AgentSuggestion{}
	}
	return &WorkflowState{
		CompanyName:        params.CompanyName,
		Slug:               params.Slug,
		StartedAt:          startedAt,
		Status:             "active",
		Contacts:           contacts,
		OutreachAttempts:   outreachAttempts,
		AgentSuggestions:   agentSuggestions,
		WorkerRestartCount: params.WorkerRestartCount,
	}
}

// WorkflowState is the complete state of a single company's outreach workflow.
type WorkflowState struct {
	CompanyName         string
	Slug                string
	StartedAt           time.Time
	Status              string // "active", "meeting_booked"
	CurrentContact      *Contact // Deprecated: kept for backward compat.
	Contacts            []Contact
	OutreachAttempts    []OutreachAttempt
	AgentSuggestions    []AgentSuggestion
	PublicID             string `json:"PublicID,omitempty"`
	AgentTaskInProgress bool
	WorkerRestartCount  int
	LastSnapshotAt      time.Time
	MeetingBookedAt     *time.Time
	MeetingNotes        string
}

// PersistWorkflowStateRequest is the input for the PersistWorkflowState activity.
type PersistWorkflowStateRequest struct {
	State *WorkflowState
	Event *ActivityEvent
}
