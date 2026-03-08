package agent

import "github.com/stephenmontague/ttm-tracker/server/internal/repository"

// AgentActivities holds dependencies for the AI agent activities.
// Registered with the Temporal worker via w.RegisterActivity().
type AgentActivities struct {
	CompanyRepo *repository.CompanyRepository
}
