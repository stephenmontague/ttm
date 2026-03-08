package agent

import (
	"fmt"
	"strings"
	"time"

	"github.com/stephenmontague/ttm-tracker/server/internal/models"
)

// SystemPrompt returns a task-specific system prompt with workflow context injected.
func SystemPrompt(taskType string, state *models.WorkflowState, contactName, contactRole string) string {
	var sb strings.Builder

	sb.WriteString("You are an AI agent assisting a BDR (Business Development Rep) with B2B outreach. ")
	sb.WriteString("You are embedded inside a Temporal workflow that tracks outreach to a specific company. ")
	sb.WriteString("You have access to tools that let you inspect the workflow state and draft messages.\n\n")

	// Inject workflow context
	daysSince := int(time.Since(state.StartedAt).Hours() / 24)
	fmt.Fprintf(&sb, "CURRENT CONTEXT:\n")
	fmt.Fprintf(&sb, "- Company: %s\n", state.CompanyName)
	fmt.Fprintf(&sb, "- Days since outreach started: %d\n", daysSince)
	fmt.Fprintf(&sb, "- Total outreach attempts: %d\n", len(state.OutreachAttempts))

	activeContacts := 0
	for _, c := range state.Contacts {
		if c.Active {
			activeContacts++
		}
	}
	fmt.Fprintf(&sb, "- Active contacts: %d\n", activeContacts)
	fmt.Fprintf(&sb, "- Workflow status: %s\n", state.Status)

	if contactName != "" {
		fmt.Fprintf(&sb, "- Selected contact: %s", contactName)
		if contactRole != "" {
			fmt.Fprintf(&sb, " (%s)", contactRole)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Task-specific instructions
	switch taskType {
	case "suggest_contact":
		sb.WriteString("YOUR TASK: Find and suggest a new person to reach out to at this company.\n\n")
		sb.WriteString("INSTRUCTIONS:\n")
		sb.WriteString("1. First call get_workflow_state to review who has already been contacted and the full outreach history.\n")
		sb.WriteString("2. Based on the outreach history, suggest a new contact who would be a good fit.\n")
		sb.WriteString("3. Consider roles like engineering leaders, infrastructure/platform engineers, or technical decision-makers.\n")
		sb.WriteString("4. Return a specific suggestion with: name (if you can determine one), role/title to target, reasoning for why they would be a good contact, and suggested approach.\n")

	case "draft_message":
		if contactName != "" {
			fmt.Fprintf(&sb, "YOUR TASK: Draft a personalized outreach message for %s at this company.\n\n", contactName)
		} else {
			sb.WriteString("YOUR TASK: Draft a personalized outreach message for a contact at this company.\n\n")
		}
		sb.WriteString("INSTRUCTIONS:\n")
		sb.WriteString("1. First call get_workflow_state to understand the full outreach history and who you're writing to.\n")
		sb.WriteString("2. Consider the outreach cadence — how many attempts have been made, through which channels, and how long it's been.\n")
		sb.WriteString("3. Draft a message that is concise, personalized, and relevant.\n")
		sb.WriteString("4. Use the draft_message tool to structure your draft with the recipient, channel, subject, and body.\n")
		sb.WriteString("5. In your final response, explain your reasoning and the personalization angles you used.\n")

	case "next_action":
		sb.WriteString("YOUR TASK: Recommend the next best outreach action for this company.\n\n")
		sb.WriteString("INSTRUCTIONS:\n")
		sb.WriteString("1. First call get_workflow_state to review the full outreach timeline.\n")
		sb.WriteString("2. Analyze the cadence: when was the last outreach? Which channels have been used? How many attempts per contact?\n")
		sb.WriteString("3. Consider timing: is it too soon to follow up? Has too much time passed?\n")
		sb.WriteString("4. Return a specific, actionable recommendation: what to do next, which channel, which contact, and when.\n")
		sb.WriteString("5. Explain your reasoning based on the outreach patterns.\n")

	default:
		sb.WriteString("YOUR TASK: Help the BDR with their outreach to this company.\n\n")
		sb.WriteString("Start by calling get_workflow_state to understand the current situation, then provide your analysis and recommendations.\n")
	}

	return sb.String()
}
