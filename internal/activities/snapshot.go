package activities

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/stephenmontague/ttm-tracker/internal/models"
	"github.com/stephenmontague/ttm-tracker/internal/repository"
)

// Activities holds dependencies for all activity implementations.
// This is the standard Temporal Go pattern for injecting dependencies
// (like DB connections) into activities. You register the struct with
// the worker, and Temporal calls its methods.
type Activities struct {
	CompanyRepo *repository.CompanyRepository
}

// SnapshotStateToCache persists the current workflow state to PostgreSQL.
// The public UI reads from this cache instead of querying Temporal directly.
func (a *Activities) SnapshotStateToCache(ctx context.Context, state *models.WorkflowState) error {
	logger := activity.GetLogger(ctx)

	now := time.Now()
	elapsedDays := int(math.Floor(now.Sub(state.StartedAt).Hours() / 24))

	var contactRole *string
	if state.CurrentContact != nil {
		contactRole = &state.CurrentContact.Role
	}

	row := &repository.CompanyRow{
		ID:                 fmt.Sprintf("outreach-%s", state.Slug),
		CompanyName:        state.CompanyName,
		Slug:               state.Slug,
		StartedAt:          state.StartedAt,
		Status:             state.Status,
		ElapsedDays:        elapsedDays,
		OutreachCount:      len(state.OutreachAttempts),
		RestartCount:       state.WorkerRestartCount,
		CurrentContactRole: contactRole,
		MeetingBookedAt:    state.MeetingBookedAt,
		LastSnapshotAt:     &now,
	}

	if err := a.CompanyRepo.UpsertWorkflow(ctx, row); err != nil {
		return fmt.Errorf("snapshot state: %w", err)
	}

	logger.Info("State snapshot saved",
		"company", state.CompanyName,
		"status", state.Status,
		"elapsed_days", elapsedDays,
		"outreach_count", len(state.OutreachAttempts),
	)

	return nil
}
