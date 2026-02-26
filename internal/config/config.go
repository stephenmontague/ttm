package config

import "os"

const (
	TaskQueue    = "ttm-tracker"
	WorkflowName = "CompanyOutreachWorkflow"
)

// Signal channel names
const (
	SignalLogOutreach   = "log_outreach"
	SignalUpdateContact = "update_contact"
	SignalRequestAgent  = "request_agent_help"
	SignalMeetingBooked = "meeting_booked"
	SignalPause         = "pause"
	SignalResume        = "resume"
)

// Query names
const (
	QueryGetState = "get_state"
)

func GetTaskQueue() string {
	if q := os.Getenv("TEMPORAL_TASK_QUEUE"); q != "" {
		return q
	}
	return TaskQueue
}

func GetAPIPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

func GetDatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}
