package agent

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	agentpkg "github.com/stephenmontague/ttm-tracker/server/internal/agent"
	"github.com/stephenmontague/ttm-tracker/server/internal/config"
)

// ExecuteAgentTool dispatches a tool call to its implementation.
// Each tool execution is a separate Temporal activity for independent retry.
func (a *AgentActivities) ExecuteAgentTool(ctx context.Context, req agentpkg.ExecuteToolRequest) (*agentpkg.ExecuteToolResponse, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing agent tool", "tool", req.ToolName)

	lushaAPIKey := config.GetLushaAPIKey()
	executor := agentpkg.NewToolExecutor(lushaAPIKey, req.State)

	result, err := executor.Execute(ctx, req.ToolName, req.ToolInput)
	if err != nil {
		logger.Warn("Tool execution failed", "tool", req.ToolName, "error", err)
		return &agentpkg.ExecuteToolResponse{
			Content: fmt.Sprintf("Tool error: %s", err.Error()),
			IsError: true,
		}, nil // Return the error as a tool result, not an activity failure
	}

	logger.Info("Tool executed successfully", "tool", req.ToolName, "resultLength", len(result))
	return &agentpkg.ExecuteToolResponse{
		Content: result,
	}, nil
}
