package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations creates the database schema if it doesn't exist.
// Called on startup from main.go.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Running database migrations...")

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Migrations complete")
	return nil
}

var migrations = []string{
	// Cached workflow state for public consumption
	`CREATE TABLE IF NOT EXISTS company_workflows (
		id                   TEXT PRIMARY KEY,
		company_name         TEXT NOT NULL,
		slug                 TEXT UNIQUE NOT NULL,
		started_at           TIMESTAMPTZ NOT NULL,
		status               TEXT NOT NULL DEFAULT 'active',
		elapsed_days         INTEGER DEFAULT 0,
		outreach_count       INTEGER DEFAULT 0,
		restart_count        INTEGER DEFAULT 0,
		current_contact_role TEXT,
		meeting_booked_at    TIMESTAMPTZ,
		last_snapshot_at     TIMESTAMPTZ,
		updated_at           TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Sanitized activity feed for public display
	`CREATE TABLE IF NOT EXISTS activity_feed (
		id          SERIAL PRIMARY KEY,
		workflow_id TEXT NOT NULL REFERENCES company_workflows(id),
		timestamp   TIMESTAMPTZ NOT NULL,
		event_type  TEXT NOT NULL,
		description TEXT NOT NULL,
		channel     TEXT,
		created_at  TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Add public_id column for opaque public-facing URLs
	`ALTER TABLE company_workflows ADD COLUMN IF NOT EXISTS public_id TEXT UNIQUE DEFAULT gen_random_uuid()::TEXT`,
	`UPDATE company_workflows SET public_id = gen_random_uuid()::TEXT WHERE public_id IS NULL`,

	// Full agent suggestions (admin only)
	`CREATE TABLE IF NOT EXISTS agent_suggestions (
		id            SERIAL PRIMARY KEY,
		workflow_id   TEXT NOT NULL REFERENCES company_workflows(id),
		task_type     TEXT NOT NULL,
		request       TEXT,
		response      TEXT,
		draft_message TEXT,
		created_at    TIMESTAMPTZ DEFAULT NOW()
	)`,

	// Multi-contact tracking
	`ALTER TABLE company_workflows ADD COLUMN IF NOT EXISTS contact_count INTEGER DEFAULT 0`,
}
