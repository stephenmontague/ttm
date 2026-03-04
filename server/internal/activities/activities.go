package activities

import "github.com/stephenmontague/ttm-tracker/server/internal/repository"

// Activities holds dependencies for all activity implementations.
// This is the standard Temporal Go pattern for injecting dependencies
// (like DB connections) into activities. You register the struct with
// the worker, and Temporal calls its methods.
type Activities struct {
	CompanyRepo *repository.CompanyRepository
}
