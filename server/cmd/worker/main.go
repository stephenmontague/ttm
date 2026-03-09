package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/stephenmontague/ttm-tracker/server/internal/activities"
	agentactivities "github.com/stephenmontague/ttm-tracker/server/internal/activities/agent"
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

	// Register agent activities (CallClaude, ExecuteAgentTool, SaveAgentSuggestion)
	agentActs := &agentactivities.AgentActivities{
		CompanyRepo: companyRepo,
	}
	w.RegisterActivity(agentActs)

	// Signal active workflows about worker restart.
	go signalActiveWorkflows(c, config.GetTaskQueue())

	log.Printf("Starting worker on task queue: %s", taskQueue)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}

// signalActiveWorkflows finds all running workflows and signals each one
// that the worker process has (re)started.
func signalActiveWorkflows(c client.Client, taskQueue string) {
	// Small delay to let the worker register with Temporal.
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := fmt.Sprintf("TaskQueue = '%s' AND ExecutionStatus = 'Running'", taskQueue)
	var nextPageToken []byte
	signaled := 0

	for {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Query:         query,
			NextPageToken: nextPageToken,
		})
		if err != nil {
			log.Printf("Failed to list workflows for restart signaling: %v", err)
			return
		}

		for _, exec := range resp.GetExecutions() {
			if exec.GetStatus() != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
				continue
			}
			wfID := exec.GetExecution().GetWorkflowId()
			err := c.SignalWorkflow(ctx, wfID, "", config.SignalWorkerRestarted, struct{}{})
			if err != nil {
				log.Printf("Failed to signal workflow %s: %v", wfID, err)
				continue
			}
			signaled++
		}

		nextPageToken = resp.GetNextPageToken()
		if len(nextPageToken) == 0 {
			break
		}
	}

	log.Printf("Worker restart: signaled %d active workflow(s)", signaled)
}
