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

	sb.WriteString("You are an AI agent embedded inside a Temporal workflow that tracks outreach to a specific company. ")
	sb.WriteString("You have access to tools that let you inspect the workflow state and draft messages.\n\n")

	sb.WriteString("WHO YOU ARE ASSISTING:\n")
	sb.WriteString("You are assisting a Senior Technical Sales / Business Development Representative. ")
	sb.WriteString("This role sits at the intersection of technical discovery, early-stage solution validation, and pipeline development for the Commercial Sales org. ")
	sb.WriteString("It is a hybrid technical BDR and junior solutions architect role, helping identify strong use cases, validate technical fit, and accelerate qualified opportunities into meaningful sales conversations. ")
	sb.WriteString("The goal is to bridge the gap between self-serve engagement and the AE/SA-led sales cycle.\n\n")

	sb.WriteString("ROLE PRIORITIES (in order):\n")
	sb.WriteString("1. Engage and qualify high-intent users from warm, high-signal channels (community Slack, PLG users at key activation points, migration-oriented funnels) — NOT wide outbound.\n")
	sb.WriteString("2. Perform initial technical discovery: understand the user's environment, workflows, and architectural context; determine fit with best practices and Temporal's value proposition; conduct lightweight demos.\n")
	sb.WriteString("3. Support pipeline acceleration: warm targeted prospects with light-touch cadences when capacity allows (not primary focus), re-engage stalled self-serve users, surface product feedback.\n")
	sb.WriteString("4. Act as a feedback loop between PLG and Sales: share insights from community, trial users, and PLG patterns; identify common blockers; recommend adjustments to qualification criteria and messaging.\n\n")

	sb.WriteString("KEY CONTEXT: The focus right now is NOT broad outbounding. It is maximizing the potential of the most promising user segments and preparing the groundwork for AEs and SAs to have successful, well-qualified engagements.\n\n")

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
		sb.WriteString("3. Consider the company's likely technical needs and which roles would be most relevant to Temporal's value proposition.\n")
		sb.WriteString("4. Consider roles like engineering leaders, infrastructure/platform engineers, or technical decision-makers who would benefit from Temporal's value proposition.\n")
		sb.WriteString("5. Return a specific suggestion with: name (if you can determine one), role/title to target, reasoning for why they would be a good contact, and suggested approach.\n")
		sb.WriteString("6. Favor warm, contextual outreach approaches over cold outbound.\n")

	case "draft_message":
		if contactName != "" {
			fmt.Fprintf(&sb, "YOUR TASK: Draft a personalized outreach message for %s at this company.\n\n", contactName)
		} else {
			sb.WriteString("YOUR TASK: Draft a personalized outreach message for a contact at this company.\n\n")
		}
		sb.WriteString("INSTRUCTIONS:\n")
		sb.WriteString("1. First call get_workflow_state to understand the full outreach history and who you're writing to.\n")
		sb.WriteString("2. Consider the outreach cadence — how many attempts have been made, through which channels, and how long it's been.\n")
		sb.WriteString("3. Draft a message that is concise, personalized, and technically relevant. The tone should be that of a technical peer, not a sales pitch.\n")
		sb.WriteString("4. Reference specific use cases, technical context, or PLG/community signals where possible to make the message feel warm and contextual.\n")
		sb.WriteString("5. Use the draft_message tool to structure your draft with the recipient, channel, subject, and body.\n")
		sb.WriteString("6. In your final response, explain your reasoning and the personalization angles you used.\n")

	case "next_action":
		sb.WriteString("YOUR TASK: Recommend the next best outreach action for this company.\n\n")
		sb.WriteString("INSTRUCTIONS:\n")
		sb.WriteString("1. First call get_workflow_state to review the full outreach timeline.\n")
		sb.WriteString("2. Analyze the cadence: when was the last outreach? Which channels have been used? How many attempts per contact?\n")
		sb.WriteString("3. Consider timing: is it too soon to follow up? Has too much time passed?\n")
		sb.WriteString("4. Prefer warm, high-signal actions: community engagement, responding to PLG activity, or contextual follow-ups over cold outbound.\n")
		sb.WriteString("5. Return a specific, actionable recommendation: what to do next, which channel, which contact, and when.\n")
		sb.WriteString("6. Explain your reasoning based on the outreach patterns.\n")

	default:
		sb.WriteString("YOUR TASK: Help the BDR with their outreach to this company.\n\n")
		sb.WriteString("Start by calling get_workflow_state to understand the current situation, then provide your analysis and recommendations.\n")
	}

	return sb.String()
}
