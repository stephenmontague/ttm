package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/stephenmontague/ttm-tracker/server/internal/activities"
	"github.com/stephenmontague/ttm-tracker/server/internal/config"
	"github.com/stephenmontague/ttm-tracker/server/internal/database"
	"github.com/stephenmontague/ttm-tracker/server/internal/repository"
	ttmtemporal "github.com/stephenmontague/ttm-tracker/server/internal/temporal"
	"github.com/stephenmontague/ttm-tracker/server/internal/workflow/outreach"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	ctx := context.Background()

	// Initialize PostgreSQL
	dbPool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	log.Println("Connected to PostgreSQL")

	if err := database.RunMigrations(ctx, dbPool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Temporal client
	c, err := ttmtemporal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()
	log.Println("Connected to Temporal")

	// Create activity struct with dependencies
	companyRepo := repository.NewCompanyRepository(dbPool)
	acts := &activities.Activities{
		CompanyRepo: companyRepo,
	}

	taskQueue := config.GetTaskQueue()
	w := worker.New(c, taskQueue, worker.Options{})

	w.RegisterWorkflowWithOptions(outreach.Workflow, workflow.RegisterOptions{
		Name: config.WorkflowName,
	})
	w.RegisterActivity(acts)

	log.Printf("Starting worker on task queue: %s", taskQueue)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
