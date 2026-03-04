package models

// ActivityEvent is raw event metadata passed from the workflow to the
// PersistWorkflowState activity. The activity sanitizes it before writing
// to the public activity_feed table.
type ActivityEvent struct {
	EventType   string // "outreach", "contact_change", "agent_action", "status_change", "meeting_booked", "workflow_started"
	Channel     string // optional, for outreach events
	Description string // raw metadata (role, task type, etc.) — activity sanitizes for public display
}
