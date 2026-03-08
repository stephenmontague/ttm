package agent

import (
	"context"

	"github.com/stephenmontague/ttm-tracker/server/internal/config"
)

// CheckLushaEnabled returns true if a Lusha API key is configured.
// This is an activity (not inline code) because workflows cannot read
// environment variables directly without breaking determinism.
func (a *AgentActivities) CheckLushaEnabled(ctx context.Context) (bool, error) {
	return config.GetLushaAPIKey() != "", nil
}
