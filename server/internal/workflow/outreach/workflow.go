package outreach

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/stephenmontague/ttm-tracker/server/internal/agent"
	"github.com/stephenmontague/ttm-tracker/server/internal/config"
	"github.com/stephenmontague/ttm-tracker/server/internal/models"
)

// persistState is a helper that executes PersistWorkflowState and logs errors
// without failing the workflow.
func persistState(ctx, actCtx workflow.Context, logger log.Logger, req models.PersistWorkflowStateRequest) {
	if err := workflow.ExecuteActivity(actCtx, "PersistWorkflowState", req).Get(ctx, nil); err != nil {
		logger.Error("Failed to persist workflow state", "error", err)
	}
}

// Workflow is a long-running workflow that tracks outreach to a single company.
// One workflow instance runs per company, potentially for weeks or months,
// until a meeting is booked.
func Workflow(ctx workflow.Context, params models.WorkflowParams) (*models.WorkflowState, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CompanyOutreachWorkflow started", "company", params.CompanyName)

	startedAt := params.StartedAt
	if startedAt.IsZero() {
		startedAt = workflow.Now(ctx) // First run — use deterministic workflow clock.
	}
	state := models.NewWorkflowState(params, startedAt)

	// Register query handler for reading live state.
	err := workflow.SetQueryHandler(ctx, config.QueryGetState, func() (*models.WorkflowState, error) {
		return state, nil
	})
	if err != nil {
		return nil, err
	}

	// Register signal channels.
	logOutreachCh := workflow.GetSignalChannel(ctx, config.SignalLogOutreach)
	addContactCh := workflow.GetSignalChannel(ctx, config.SignalAddContact)
	removeContactCh := workflow.GetSignalChannel(ctx, config.SignalRemoveContact)
	updateContactCh := workflow.GetSignalChannel(ctx, config.SignalUpdateContact) // Deprecated: backward compat.
	agentHelpCh := workflow.GetSignalChannel(ctx, config.SignalRequestAgent)
	meetingBookedCh := workflow.GetSignalChannel(ctx, config.SignalMeetingBooked)
	workerRestartCh := workflow.GetSignalChannel(ctx, config.SignalWorkerRestarted)

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

	// Initial persist so the company appears in the dashboard immediately.
	persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state, Event: &models.ActivityEvent{EventType: "workflow_started"}})
	state.LastSnapshotAt = workflow.Now(ctx)

	// Create one timer that we cancel and recreate only when it fires.
	// This prevents orphaned timers accumulating across loop iterations.
	timerCtx, timerCancel := workflow.WithCancel(ctx)
	timerFuture := workflow.NewTimer(timerCtx, 24*time.Hour)

	// Main event loop — waits for signals or a daily timer tick.
	eventCount := 0
	for state.Status != "meeting_booked" {
		timerFired := false
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(logOutreachCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.LogOutreachSignal
			ch.Receive(ctx, &signal)

			// Use explicit contact name from signal; fall back to CurrentContact for old signals.
			contactName := signal.ContactName
			if contactName == "" && state.CurrentContact != nil {
				contactName = state.CurrentContact.Name
			}

			state.OutreachAttempts = append(state.OutreachAttempts, models.OutreachAttempt{
				Timestamp: workflow.Now(ctx),
				Channel:   signal.Channel,
				Notes:     signal.Notes,
				Contact:   contactName,
			})

			logger.Info("Outreach logged", "channel", signal.Channel, "contact", contactName, "total_attempts", len(state.OutreachAttempts))
			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state, Event: &models.ActivityEvent{
				EventType: "outreach",
				Channel:   signal.Channel,
			}})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddReceive(addContactCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.AddContactSignal
			ch.Receive(ctx, &signal)

			found := false
			for i, c := range state.Contacts {
				if c.Name == signal.Name {
					state.Contacts[i].Role = signal.Role
					state.Contacts[i].LinkedIn = signal.LinkedIn
					state.Contacts[i].Active = true
					found = true
					break
				}
			}
			if !found {
				state.Contacts = append(state.Contacts, models.Contact{
					Name:     signal.Name,
					Role:     signal.Role,
					LinkedIn: signal.LinkedIn,
					Active:   true,
					AddedAt:  workflow.Now(ctx),
				})
			}

			logger.Info("Contact added", "name", signal.Name, "role", signal.Role, "total_contacts", len(state.Contacts))
			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state, Event: &models.ActivityEvent{
				EventType:   "contact_change",
				Description: signal.Role,
			}})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddReceive(removeContactCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.RemoveContactSignal
			ch.Receive(ctx, &signal)

			for i, c := range state.Contacts {
				if c.Name == signal.Name {
					state.Contacts[i].Active = false
					break
				}
			}

			logger.Info("Contact deactivated", "name", signal.Name)
			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state, Event: &models.ActivityEvent{
				EventType:   "contact_change",
				Description: "contact removed",
			}})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		// Deprecated: backward compat for in-flight workflows using the old update_contact signal.
		selector.AddReceive(updateContactCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.UpdateContactSignal
			ch.Receive(ctx, &signal)

			// Treat as add-contact for migration.
			found := false
			for i, c := range state.Contacts {
				if c.Name == signal.Name {
					state.Contacts[i].Role = signal.Role
					state.Contacts[i].LinkedIn = signal.LinkedIn
					state.Contacts[i].Active = true
					found = true
					break
				}
			}
			if !found {
				state.Contacts = append(state.Contacts, models.Contact{
					Name:     signal.Name,
					Role:     signal.Role,
					LinkedIn: signal.LinkedIn,
					Active:   true,
					AddedAt:  workflow.Now(ctx),
				})
			}

			// Also set CurrentContact for backward compat.
			state.CurrentContact = &models.Contact{
				Name:     signal.Name,
				Role:     signal.Role,
				LinkedIn: signal.LinkedIn,
				Active:   true,
				AddedAt:  workflow.Now(ctx),
			}

			logger.Info("Contact updated (legacy signal)", "name", signal.Name, "role", signal.Role)
			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state, Event: &models.ActivityEvent{
				EventType:   "contact_change",
				Description: signal.Role,
			}})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddReceive(agentHelpCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.RequestAgentHelpSignal
			ch.Receive(ctx, &signal)

			logger.Info("Agent help requested", "taskType", signal.TaskType, "contact", signal.ContactName)
			state.AgentTaskInProgress = true
			suggestion := runAgentLoop(ctx, logger, state, signal)
			state.AgentTaskInProgress = false
			state.AgentSuggestions = append(state.AgentSuggestions, suggestion)

			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state, Event: &models.ActivityEvent{
				EventType:   "agent_action",
				Description: signal.TaskType,
			}})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddReceive(meetingBookedCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.MeetingBookedSignal
			ch.Receive(ctx, &signal)

			state.Status = "meeting_booked"
			state.MeetingBookedAt = &signal.Date
			state.MeetingNotes = signal.Notes

			logger.Info("Meeting booked!", "date", signal.Date)
			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state, Event: &models.ActivityEvent{EventType: "meeting_booked"}})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddReceive(workerRestartCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal struct{}
			ch.Receive(ctx, &signal)
			state.WorkerRestartCount++
			logger.Info("Worker restart detected", "restart_count", state.WorkerRestartCount)
			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{
				State: state,
				Event: &models.ActivityEvent{EventType: "worker_restart"},
			})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			_ = f.Get(ctx, nil)
			timerFired = true
			logger.Info("Daily timer tick")
			persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state})
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.Select(ctx)
		eventCount++

		// Only create a new timer if the old one fired.
		// If a signal woke us up, the existing timer keeps ticking.
		if timerFired {
			timerCtx, timerCancel = workflow.WithCancel(ctx)
			timerFuture = workflow.NewTimer(timerCtx, 24*time.Hour)
		}

		// Continue-as-new to prevent unbounded event history growth.
		if eventCount >= 1000 {
			timerCancel()
			logger.Info("Continuing as new", "events", eventCount)
			return nil, workflow.NewContinueAsNewError(ctx, Workflow, models.WorkflowParams{
				CompanyName:        state.CompanyName,
				Slug:               state.Slug,
				StartedAt:          state.StartedAt,
				Contacts:           state.Contacts,
				OutreachAttempts:   state.OutreachAttempts,
				AgentSuggestions:   state.AgentSuggestions,
				WorkerRestartCount: state.WorkerRestartCount,
			})
		}
	}

	timerCancel()

	persistState(ctx, actCtx, logger, models.PersistWorkflowStateRequest{State: state})
	state.LastSnapshotAt = workflow.Now(ctx)
	logger.Info("Workflow completed", "company", state.CompanyName)
	return state, nil
}

// runAgentLoop orchestrates the agentic loop as workflow logic.
// Each Claude API call and tool execution is a separate Temporal activity.
func runAgentLoop(ctx workflow.Context, logger log.Logger, state *models.WorkflowState, signal models.RequestAgentHelpSignal) models.AgentSuggestion {
	// Activity options for Claude API calls (longer timeout).
	claudeActCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 90 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
	})

	// Activity options for tool execution (shorter timeout).
	toolActCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
	})

	// Activity options for saving the suggestion.
	saveActCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	})

	// Check if Lusha API key is configured (activity because workflows can't read env vars).
	var lushaEnabled bool
	if err := workflow.ExecuteActivity(saveActCtx, "CheckLushaEnabled").Get(ctx, &lushaEnabled); err != nil {
		logger.Error("Failed to check Lusha status", "error", err)
	}
	tools := agent.ToolDefinitions(lushaEnabled)

	// Find the selected contact's role for prompt context.
	var contactRole string
	if signal.ContactName != "" {
		for _, c := range state.Contacts {
			if c.Name == signal.ContactName {
				contactRole = c.Role
				break
			}
		}
	}

	systemPrompt := agent.SystemPrompt(signal.TaskType, state, signal.ContactName, contactRole)
	messages := agent.BuildInitialMessages(signal.TaskType, signal.Context, signal.ContactName)

	const maxIterations = 10

	for i := range maxIterations {
		// Call Claude activity.
		var claudeResp agent.CallClaudeResponse
		err := workflow.ExecuteActivity(claudeActCtx, "CallClaude", agent.CallClaudeRequest{
			System:   systemPrompt,
			Messages: messages,
			Tools:    tools,
		}).Get(ctx, &claudeResp)
		if err != nil {
			logger.Error("CallClaude activity failed", "error", err, "iteration", i)
			return models.AgentSuggestion{
				Timestamp: workflow.Now(ctx),
				TaskType:  signal.TaskType,
				Request:   signal.Context,
				Response:  fmt.Sprintf("Agent failed: %s", err.Error()),
			}
		}

		// Append assistant response to conversation.
		messages = append(messages, agent.Message{
			Role:    "assistant",
			Content: claudeResp.Content,
		})

		// If Claude is done, break.
		if claudeResp.StopReason == "end_turn" {
			break
		}

		// Process tool_use blocks.
		var toolResults []agent.ContentBlock
		for _, block := range claudeResp.Content {
			if block.Type != "tool_use" {
				continue
			}

			logger.Info("Agent calling tool", "tool", block.Name, "iteration", i)

			var toolResp agent.ExecuteToolResponse
			err := workflow.ExecuteActivity(toolActCtx, "ExecuteAgentTool", agent.ExecuteToolRequest{
				ToolName:  block.Name,
				ToolInput: block.Input,
				State:     state,
			}).Get(ctx, &toolResp)
			if err != nil {
				// Activity itself failed (not a tool error) — return error to Claude.
				toolResults = append(toolResults, agent.ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   fmt.Sprintf("Activity error: %s", err.Error()),
					IsError:   true,
				})
				continue
			}

			toolResults = append(toolResults, agent.ContentBlock{
				Type:      "tool_result",
				ToolUseID: block.ID,
				Content:   toolResp.Content,
				IsError:   toolResp.IsError,
			})
		}

		// Append tool results as a user message.
		messages = append(messages, agent.Message{
			Role:    "user",
			Content: toolResults,
		})

		// If we're at the iteration limit, force Claude to wrap up.
		if i == maxIterations-2 {
			messages = append(messages, agent.Message{
				Role: "user",
				Content: []agent.ContentBlock{
					{Type: "text", Text: "You have reached the tool call limit. Please provide your final answer now."},
				},
			})
		}
	}

	// Build the suggestion from the final response.
	finalText := agent.ExtractFinalText(messages[len(messages)-1].Content)
	if messages[len(messages)-1].Role != "assistant" && len(messages) >= 2 {
		// Last message might be tool results; check the one before.
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "assistant" {
				finalText = agent.ExtractFinalText(messages[i].Content)
				break
			}
		}
	}

	draftMessage := agent.ExtractDraftMessage(messages)

	suggestion := models.AgentSuggestion{
		Timestamp:    workflow.Now(ctx),
		TaskType:     signal.TaskType,
		ContactName:  signal.ContactName,
		Request:      signal.Context,
		Response:     finalText,
		DraftMessage: draftMessage,
	}

	// Persist to database.
	workflowID := fmt.Sprintf("outreach-%s", state.Slug)
	if err := workflow.ExecuteActivity(saveActCtx, "SaveAgentSuggestion", agent.SaveSuggestionRequest{
		WorkflowID: workflowID,
		Suggestion: suggestion,
	}).Get(ctx, nil); err != nil {
		logger.Error("Failed to save agent suggestion", "error", err)
	}

	return suggestion
}
