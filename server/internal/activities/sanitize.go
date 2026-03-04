package activities

import "github.com/stephenmontague/ttm-tracker/server/internal/models"

// sanitizeEvent converts raw workflow event metadata into a public-safe
// activity feed description. Contact names and notes are stripped.
func sanitizeEvent(event *models.ActivityEvent) string {
	switch event.EventType {
	case "outreach":
		if event.Channel != "" {
			return "Outreach via " + event.Channel
		}
		return "Outreach logged"
	case "contact_change":
		if event.Description == "contact removed" {
			return "Contact deactivated"
		}
		if event.Description != "" {
			return "New contact identified (" + event.Description + ")"
		}
		return "Contact updated"
	case "agent_action":
		if event.Description != "" {
			return "AI agent " + event.Description + " requested"
		}
		return "AI agent task requested"
	case "status_change":
		return "Workflow " + event.Description
	case "meeting_booked":
		return "Meeting booked!"
	case "workflow_started":
		return "Outreach tracking started"
	default:
		return "Activity recorded"
	}
}
