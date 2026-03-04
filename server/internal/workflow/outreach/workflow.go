package outreach

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/stephenmontague/ttm-tracker/server/internal/config"
	"github.com/stephenmontague/ttm-tracker/server/internal/models"
)

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

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

	// Initial persist so the company appears in the dashboard immediately.
	_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, &models.ActivityEvent{EventType: "workflow_started"}).Get(ctx, nil)
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
			_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, &models.ActivityEvent{
				EventType: "outreach",
				Channel:   signal.Channel,
			}).Get(ctx, nil)
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
			_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, &models.ActivityEvent{
				EventType:   "contact_change",
				Description: signal.Role,
			}).Get(ctx, nil)
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
			_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, &models.ActivityEvent{
				EventType:   "contact_change",
				Description: "contact removed",
			}).Get(ctx, nil)
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
			_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, &models.ActivityEvent{
				EventType:   "contact_change",
				Description: signal.Role,
			}).Get(ctx, nil)
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddReceive(agentHelpCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.RequestAgentHelpSignal
			ch.Receive(ctx, &signal)

			// TODO (Phase 3): Execute RunAgent activity here
			logger.Info("Agent help requested", "taskType", signal.TaskType)
			_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, &models.ActivityEvent{
				EventType:   "agent_action",
				Description: signal.TaskType,
			}).Get(ctx, nil)
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddReceive(meetingBookedCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.MeetingBookedSignal
			ch.Receive(ctx, &signal)

			state.Status = "meeting_booked"
			state.MeetingBookedAt = &signal.Date
			state.MeetingNotes = signal.Notes

			logger.Info("Meeting booked!", "date", signal.Date)
			_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, &models.ActivityEvent{EventType: "meeting_booked"}).Get(ctx, nil)
			state.LastSnapshotAt = workflow.Now(ctx)
		})

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			_ = f.Get(ctx, nil)
			timerFired = true
			logger.Info("Daily timer tick")
			_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, nil).Get(ctx, nil)
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
				CompanyName: state.CompanyName,
				Slug:        state.Slug,
				StartedAt:   state.StartedAt,
				Contacts:    state.Contacts,
			})
		}
	}

	timerCancel()

	_ = workflow.ExecuteActivity(actCtx, "PersistWorkflowState", state, nil).Get(ctx, nil)
	state.LastSnapshotAt = workflow.Now(ctx)
	logger.Info("Workflow completed", "company", state.CompanyName)
	return state, nil
}
