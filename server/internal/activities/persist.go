package activities

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/stephenmontague/ttm-tracker/server/internal/models"
	"github.com/stephenmontague/ttm-tracker/server/internal/repository"
)

// PersistWorkflowState persists the current workflow state to PostgreSQL and
// optionally writes a sanitized event to the public activity feed — all in
// a single database transaction. The public UI reads from these tables instead
// of querying Temporal directly.
func (a *Activities) PersistWorkflowState(ctx context.Context, req models.PersistWorkflowStateRequest) error {
	logger := activity.GetLogger(ctx)

	state := req.State
	event := req.Event

	now := time.Now()
	elapsedDays := int(math.Floor(now.Sub(state.StartedAt).Hours() / 24))

	// Count active contacts for the public-facing contact_count.
	activeContactCount := 0
	var contactRole *string
	if len(state.Contacts) > 0 {
		for _, c := range state.Contacts {
			if c.Active {
				activeContactCount++
				if contactRole == nil {
					role := c.Role
					contactRole = &role
				}
			}
		}
	} else if state.CurrentContact != nil {
		// Backward compat: old workflow state with only CurrentContact.
		contactRole = &state.CurrentContact.Role
		activeContactCount = 1
	}

	workflowID := fmt.Sprintf("outreach-%s", state.Slug)

	row := &repository.CompanyRow{
		ID:                 workflowID,
		CompanyName:        state.CompanyName,
		Slug:               state.Slug,
		StartedAt:          state.StartedAt,
		Status:             state.Status,
		ElapsedDays:        elapsedDays,
		OutreachCount:      len(state.OutreachAttempts),
		ContactCount:       activeContactCount,
		RestartCount:       state.WorkerRestartCount,
		CurrentContactRole: contactRole,
		MeetingBookedAt:    state.MeetingBookedAt,
		LastSnapshotAt:     &now,
	}

	var feedRow *repository.ActivityFeedRow
	if event != nil {
		feedRow = &repository.ActivityFeedRow{
			WorkflowID:  workflowID,
			Timestamp:   now,
			EventType:   event.EventType,
			Description: sanitizeEvent(event),
		}
		if event.Channel != "" {
			feedRow.Channel = &event.Channel
		}
	}

	if err := a.CompanyRepo.PersistStateAndEvent(ctx, row, feedRow); err != nil {
		return fmt.Errorf("persist workflow state: %w", err)
	}

	logger.Info("State persisted",
		"company", state.CompanyName,
		"status", state.Status,
		"elapsed_days", elapsedDays,
		"outreach_count", len(state.OutreachAttempts),
		"has_event", event != nil,
	)

	return nil
}

// SnapshotStateToCache is a backward-compatibility wrapper for in-flight
// workflows whose Temporal event history references the old activity name.
// Remove after all running workflows have continued-as-new.
func (a *Activities) SnapshotStateToCache(ctx context.Context, req models.PersistWorkflowStateRequest) error {
	return a.PersistWorkflowState(ctx, req)
}
