package agent

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	agentpkg "github.com/stephenmontague/ttm-tracker/server/internal/agent"
)

// SaveAgentSuggestion persists an agent suggestion to the database.
func (a *AgentActivities) SaveAgentSuggestion(ctx context.Context, req agentpkg.SaveSuggestionRequest) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Saving agent suggestion", "workflowID", req.WorkflowID, "taskType", req.Suggestion.TaskType)

	err := a.CompanyRepo.InsertAgentSuggestion(ctx,
		req.WorkflowID,
		req.Suggestion.TaskType,
		req.Suggestion.ContactName,
		req.Suggestion.Request,
		req.Suggestion.Response,
		req.Suggestion.DraftMessage,
		req.Suggestion.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("save agent suggestion: %w", err)
	}

	logger.Info("Agent suggestion saved")
	return nil
}
