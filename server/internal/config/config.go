package config

import (
	"os"
	"strconv"
)

const (
	TaskQueue    = "ttm-tracker"
	WorkflowName = "CompanyOutreachWorkflow"
)

// Signal channel names
const (
	SignalLogOutreach   = "log_outreach"
	SignalUpdateContact = "update_contact" // Deprecated: kept for in-flight workflow compat.
	SignalAddContact    = "add_contact"
	SignalRemoveContact = "remove_contact"
	SignalRequestAgent  = "request_agent_help"
	SignalMeetingBooked = "meeting_booked"
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

// Session auth constants
const (
	DefaultSessionCookieName = "ttm_session"
	DefaultSessionMaxAge     = 604800 // 7 days in seconds
)

func GetAdminSeedEmail() string {
	return os.Getenv("ADMIN_SEED_EMAIL")
}

func GetAdminSeedPassword() string {
	return os.Getenv("ADMIN_SEED_PASSWORD")
}

func GetSessionCookieName() string {
	if n := os.Getenv("SESSION_COOKIE_NAME"); n != "" {
		return n
	}
	return DefaultSessionCookieName
}

// AI Agent config

func GetAnthropicAPIKey() string {
	return os.Getenv("ANTHROPIC_API_KEY")
}

func GetClaudeModel() string {
	if m := os.Getenv("CLAUDE_MODEL"); m != "" {
		return m
	}
	return "claude-sonnet-4-20250514"
}

func GetLushaAPIKey() string {
	return os.Getenv("LUSHA_API_KEY")
}

func GetSessionMaxAge() int {
	if s := os.Getenv("SESSION_MAX_AGE"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return v
		}
	}
	return DefaultSessionMaxAge
}
