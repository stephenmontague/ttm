package agent

import (
	"encoding/json"
	"fmt"

	"github.com/stephenmontague/ttm-tracker/server/internal/models"
)

// CallClaudeRequest is the input for the CallClaude activity.
type CallClaudeRequest struct {
	System   string    `json:"system"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools"`
}

// CallClaudeResponse is the output of the CallClaude activity.
type CallClaudeResponse struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
}

// ExecuteToolRequest is the input for the ExecuteAgentTool activity.
type ExecuteToolRequest struct {
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
	State     *models.WorkflowState `json:"state"`
}

// ExecuteToolResponse is the output of the ExecuteAgentTool activity.
type ExecuteToolResponse struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// SaveSuggestionRequest is the input for the SaveAgentSuggestion activity.
type SaveSuggestionRequest struct {
	WorkflowID string                `json:"workflow_id"`
	Suggestion models.AgentSuggestion `json:"suggestion"`
}

// BuildInitialMessages creates the first user message for the agentic loop.
func BuildInitialMessages(taskType, userContext, contactName string) []Message {
	text := "Please help with this task."
	if userContext != "" {
		text = userContext
	} else {
		switch taskType {
		case "suggest_contact":
			text = "Please find and suggest a new contact to reach out to at this company."
		case "draft_message":
			if contactName != "" {
				text = fmt.Sprintf("Please draft a personalized outreach message for %s at this company.", contactName)
			} else {
				text = "Please draft a personalized outreach message for this company."
			}
		case "next_action":
			text = "Please recommend the best next outreach action for this company."
		}
	}

	return []Message{
		{
			Role: "user",
			Content: []ContentBlock{
				{Type: "text", Text: text},
			},
		},
	}
}

// ExtractFinalText extracts the final text response from Claude's content blocks.
func ExtractFinalText(content []ContentBlock) string {
	for _, block := range content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

// ExtractDraftMessage scans the conversation history for the last draft_message
// tool call and extracts the body field from its input.
func ExtractDraftMessage(messages []Message) string {
	// Walk backwards through messages to find the last draft_message tool call
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role != "assistant" {
			continue
		}
		for _, block := range msg.Content {
			if block.Type == "tool_use" && block.Name == "draft_message" {
				var draft struct {
					Body string `json:"body"`
				}
				if err := json.Unmarshal(block.Input, &draft); err == nil && draft.Body != "" {
					return draft.Body
				}
			}
		}
	}
	return ""
}
