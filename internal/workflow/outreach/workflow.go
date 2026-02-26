package outreach

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/stephenmontague/ttm-tracker/internal/config"
	"github.com/stephenmontague/ttm-tracker/internal/models"
)

// Workflow is a long-running workflow that tracks outreach to a single company.
// One workflow instance runs per company, potentially for weeks or months,
// until a meeting is booked.
func Workflow(ctx workflow.Context, params models.WorkflowParams) (*models.WorkflowState, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("CompanyOutreachWorkflow started", "company", params.CompanyName)

	state := initState(params)

	// Register query handler for reading live state.
	err := workflow.SetQueryHandler(ctx, config.QueryGetState, func() (*models.WorkflowState, error) {
		return state, nil
	})
	if err != nil {
		return nil, err
	}

	// Register signal channels.
	logOutreachCh := workflow.GetSignalChannel(ctx, config.SignalLogOutreach)
	updateContactCh := workflow.GetSignalChannel(ctx, config.SignalUpdateContact)
	agentHelpCh := workflow.GetSignalChannel(ctx, config.SignalRequestAgent)
	meetingBookedCh := workflow.GetSignalChannel(ctx, config.SignalMeetingBooked)
	pauseCh := workflow.GetSignalChannel(ctx, config.SignalPause)
	resumeCh := workflow.GetSignalChannel(ctx, config.SignalResume)

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

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

			contactName := ""
			if state.CurrentContact != nil {
				contactName = state.CurrentContact.Name
			}

			state.OutreachAttempts = append(state.OutreachAttempts, models.OutreachAttempt{
				Timestamp: workflow.Now(ctx),
				Channel:   signal.Channel,
				Notes:     signal.Notes,
				Contact:   contactName,
			})

			logger.Info("Outreach logged", "channel", signal.Channel, "total_attempts", len(state.OutreachAttempts))
			snapshot(actCtx, state)
		})

		selector.AddReceive(updateContactCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.UpdateContactSignal
			ch.Receive(ctx, &signal)

			state.CurrentContact = &models.Contact{
				Name:     signal.Name,
				Role:     signal.Role,
				LinkedIn: signal.LinkedIn,
			}

			logger.Info("Contact updated", "name", signal.Name, "role", signal.Role)
			snapshot(actCtx, state)
		})

		selector.AddReceive(agentHelpCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.RequestAgentHelpSignal
			ch.Receive(ctx, &signal)

			// TODO (Phase 3): Execute RunAgent activity here
			logger.Info("Agent help requested", "taskType", signal.TaskType)
			snapshot(actCtx, state)
		})

		selector.AddReceive(meetingBookedCh, func(ch workflow.ReceiveChannel, more bool) {
			var signal models.MeetingBookedSignal
			ch.Receive(ctx, &signal)

			state.Status = "meeting_booked"
			state.MeetingBookedAt = &signal.Date
			state.MeetingNotes = signal.Notes

			logger.Info("Meeting booked!", "date", signal.Date)
			snapshot(actCtx, state)
		})

		selector.AddReceive(pauseCh, func(ch workflow.ReceiveChannel, more bool) {
			ch.Receive(ctx, nil)

			state.Status = "paused"
			logger.Info("Workflow paused")
			snapshot(actCtx, state)

			// Block until resume signal arrives.
			resumeCh.Receive(ctx, nil)
			state.Status = "active"
			logger.Info("Workflow resumed")
			snapshot(actCtx, state)
		})

		selector.AddFuture(timerFuture, func(f workflow.Future) {
			_ = f.Get(ctx, nil)
			timerFired = true
			logger.Info("Daily timer tick")
			snapshot(actCtx, state)
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
			})
		}
	}

	timerCancel()

	snapshot(actCtx, state)
	logger.Info("Workflow completed", "company", state.CompanyName)
	return state, nil
}

func initState(params models.WorkflowParams) *models.WorkflowState {
	return &models.WorkflowState{
		CompanyName:      params.CompanyName,
		Slug:             params.Slug,
		StartedAt:        time.Now(),
		Status:           "active",
		OutreachAttempts: []models.OutreachAttempt{},
		AgentSuggestions: []models.AgentSuggestion{},
	}
}

func snapshot(ctx workflow.Context, state *models.WorkflowState) {
	// Reference the activity by name. When the worker registers the Activities
	// struct, Temporal registers each method by its name ("SnapshotStateToCache").
	// The workflow doesn't need a direct import — Temporal resolves it at runtime.
	_ = workflow.ExecuteActivity(ctx, "SnapshotStateToCache", state).Get(ctx, nil)
	state.LastSnapshotAt = workflow.Now(ctx)
}
