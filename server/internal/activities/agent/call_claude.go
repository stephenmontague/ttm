package agent

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	agentpkg "github.com/stephenmontague/ttm-tracker/server/internal/agent"
	"github.com/stephenmontague/ttm-tracker/server/internal/config"
)

// CallClaude sends messages to the Claude API and returns the response.
// Each call is a separate Temporal activity for independent retry and visibility.
func (a *AgentActivities) CallClaude(ctx context.Context, req agentpkg.CallClaudeRequest) (*agentpkg.CallClaudeResponse, error) {
	logger := activity.GetLogger(ctx)

	apiKey := config.GetAnthropicAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY not configured")
	}

	model := config.GetClaudeModel()
	client := agentpkg.NewClient(apiKey, model)

	logger.Info("Calling Claude API", "model", model, "messageCount", len(req.Messages))

	resp, err := client.SendMessages(ctx, &agentpkg.MessagesRequest{
		System:   req.System,
		Messages: req.Messages,
		Tools:    req.Tools,
	})
	if err != nil {
		return nil, fmt.Errorf("claude API call failed: %w", err)
	}

	logger.Info("Claude API response", "stopReason", resp.StopReason,
		"inputTokens", resp.Usage.InputTokens, "outputTokens", resp.Usage.OutputTokens)

	return &agentpkg.CallClaudeResponse{
		Content:    resp.Content,
		StopReason: resp.StopReason,
	}, nil
}
