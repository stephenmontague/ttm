package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CompanyRow represents a row in the company_workflows table.
type CompanyRow struct {
	ID                 string
	CompanyName        string
	Slug               string
	PublicID           string
	StartedAt          time.Time
	Status             string
	ElapsedDays        int
	OutreachCount      int
	ContactCount       int
	RestartCount       int
	CurrentContactRole *string
	MeetingBookedAt    *time.Time
	LastSnapshotAt     *time.Time
	UpdatedAt          time.Time
}

// ActivityFeedRow represents a row in the activity_feed table.
type ActivityFeedRow struct {
	ID          int
	WorkflowID  string
	Timestamp   time.Time
	EventType   string
	Description string
	Channel     *string
	CreatedAt   time.Time
}

type CompanyRepository struct {
	pool *pgxpool.Pool
}

func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{pool: pool}
}

// UpsertWorkflow inserts or updates the cached workflow state.
// Called by the snapshot activity whenever workflow state changes.
func (r *CompanyRepository) UpsertWorkflow(ctx context.Context, row *CompanyRow) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO company_workflows (id, company_name, slug, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			elapsed_days = EXCLUDED.elapsed_days,
			outreach_count = EXCLUDED.outreach_count,
			contact_count = EXCLUDED.contact_count,
			restart_count = EXCLUDED.restart_count,
			current_contact_role = EXCLUDED.current_contact_role,
			meeting_booked_at = EXCLUDED.meeting_booked_at,
			last_snapshot_at = EXCLUDED.last_snapshot_at,
			updated_at = NOW()
	`, row.ID, row.CompanyName, row.Slug, row.StartedAt, row.Status,
		row.ElapsedDays, row.OutreachCount, row.ContactCount, row.RestartCount,
		row.CurrentContactRole, row.MeetingBookedAt, row.LastSnapshotAt)
	if err != nil {
		return fmt.Errorf("upsert workflow: %w", err)
	}
	return nil
}

// InsertActivityFeed adds a sanitized event to the public activity feed.
func (r *CompanyRepository) InsertActivityFeed(ctx context.Context, row *ActivityFeedRow) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO activity_feed (workflow_id, timestamp, event_type, description, channel)
		VALUES ($1, $2, $3, $4, $5)
	`, row.WorkflowID, row.Timestamp, row.EventType, row.Description, row.Channel)
	if err != nil {
		return fmt.Errorf("insert activity feed: %w", err)
	}
	return nil
}

// PersistStateAndEvent upserts workflow state and optionally inserts an
// activity feed row, all within a single database transaction.
func (r *CompanyRepository) PersistStateAndEvent(ctx context.Context, row *CompanyRow, event *ActivityFeedRow) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO company_workflows (id, company_name, slug, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			elapsed_days = EXCLUDED.elapsed_days,
			outreach_count = EXCLUDED.outreach_count,
			contact_count = EXCLUDED.contact_count,
			restart_count = EXCLUDED.restart_count,
			current_contact_role = EXCLUDED.current_contact_role,
			meeting_booked_at = EXCLUDED.meeting_booked_at,
			last_snapshot_at = EXCLUDED.last_snapshot_at,
			updated_at = NOW()
	`, row.ID, row.CompanyName, row.Slug, row.StartedAt, row.Status,
		row.ElapsedDays, row.OutreachCount, row.ContactCount, row.RestartCount,
		row.CurrentContactRole, row.MeetingBookedAt, row.LastSnapshotAt)
	if err != nil {
		return fmt.Errorf("upsert workflow: %w", err)
	}

	if event != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO activity_feed (workflow_id, timestamp, event_type, description, channel)
			VALUES ($1, $2, $3, $4, $5)
		`, event.WorkflowID, event.Timestamp, event.EventType, event.Description, event.Channel)
		if err != nil {
			return fmt.Errorf("insert activity feed: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// ListCompanies returns all tracked companies (for the public dashboard).
func (r *CompanyRepository) ListCompanies(ctx context.Context) ([]CompanyRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, company_name, slug, public_id, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, updated_at
		FROM company_workflows
		ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	defer rows.Close()

	var companies []CompanyRow
	for rows.Next() {
		var c CompanyRow
		if err := rows.Scan(&c.ID, &c.CompanyName, &c.Slug, &c.PublicID, &c.StartedAt, &c.Status,
			&c.ElapsedDays, &c.OutreachCount, &c.ContactCount, &c.RestartCount,
			&c.CurrentContactRole, &c.MeetingBookedAt, &c.LastSnapshotAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan company: %w", err)
		}
		companies = append(companies, c)
	}
	return companies, nil
}

// UpdateStatus sets the status of a workflow in the cache.
func (r *CompanyRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE company_workflows SET status = $1, updated_at = NOW() WHERE id = $2
	`, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// GetCompanyBySlug returns a single company by its URL slug.
func (r *CompanyRepository) GetCompanyBySlug(ctx context.Context, slug string) (*CompanyRow, error) {
	var c CompanyRow
	err := r.pool.QueryRow(ctx, `
		SELECT id, company_name, slug, public_id, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, updated_at
		FROM company_workflows
		WHERE slug = $1
	`, slug).Scan(&c.ID, &c.CompanyName, &c.Slug, &c.PublicID, &c.StartedAt, &c.Status,
		&c.ElapsedDays, &c.OutreachCount, &c.ContactCount, &c.RestartCount,
		&c.CurrentContactRole, &c.MeetingBookedAt, &c.LastSnapshotAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get company by slug: %w", err)
	}
	return &c, nil
}

// GetCompanyByPublicID returns a single company by its public-facing UUID.
func (r *CompanyRepository) GetCompanyByPublicID(ctx context.Context, publicID string) (*CompanyRow, error) {
	var c CompanyRow
	err := r.pool.QueryRow(ctx, `
		SELECT id, company_name, slug, public_id, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, updated_at
		FROM company_workflows
		WHERE public_id = $1
	`, publicID).Scan(&c.ID, &c.CompanyName, &c.Slug, &c.PublicID, &c.StartedAt, &c.Status,
		&c.ElapsedDays, &c.OutreachCount, &c.ContactCount, &c.RestartCount,
		&c.CurrentContactRole, &c.MeetingBookedAt, &c.LastSnapshotAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get company by public id: %w", err)
	}
	return &c, nil
}

// InsertAgentSuggestion persists an agent suggestion to the database.
func (r *CompanyRepository) InsertAgentSuggestion(ctx context.Context, workflowID, taskType, request, response, draftMessage string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO agent_suggestions (workflow_id, task_type, request, response, draft_message)
		VALUES ($1, $2, $3, $4, $5)
	`, workflowID, taskType, request, response, draftMessage)
	if err != nil {
		return fmt.Errorf("insert agent suggestion: %w", err)
	}
	return nil
}

// GetActivityFeed returns the sanitized activity feed for a company.
func (r *CompanyRepository) GetActivityFeed(ctx context.Context, workflowID string) ([]ActivityFeedRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workflow_id, timestamp, event_type, description, channel, created_at
		FROM activity_feed
		WHERE workflow_id = $1
		ORDER BY timestamp DESC
		LIMIT 50
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get activity feed: %w", err)
	}
	defer rows.Close()

	var feed []ActivityFeedRow
	for rows.Next() {
		var f ActivityFeedRow
		if err := rows.Scan(&f.ID, &f.WorkflowID, &f.Timestamp, &f.EventType,
			&f.Description, &f.Channel, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan activity feed: %w", err)
		}
		feed = append(feed, f)
	}
	return feed, nil
}
