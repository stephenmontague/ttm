package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CompanyRow represents a row in the company_workflows table.
type CompanyRow struct {
	ID                  string
	CompanyName         string
	Slug                string
	PublicID            string
	StartedAt           time.Time
	Status              string
	ElapsedDays         int
	OutreachCount       int
	ContactCount        int
	RestartCount        int
	CurrentContactRole  *string
	MeetingBookedAt     *time.Time
	LastSnapshotAt      *time.Time
	UpdatedAt           time.Time
	AgentTaskInProgress bool
	MeetingNotes        string
}

// ContactRow represents a row in the contacts table.
type ContactRow struct {
	ID         int
	WorkflowID string
	Name       string
	Role       string
	LinkedIn   string
	Active     bool
	AddedAt    time.Time
}

// OutreachAttemptRow represents a row in the outreach_attempts table.
type OutreachAttemptRow struct {
	ID         int
	WorkflowID string
	Timestamp  time.Time
	Channel    string
	Notes      string
	Contact    string
}

// AgentSuggestionRow represents a row in the agent_suggestions table.
type AgentSuggestionRow struct {
	ID           int
	WorkflowID   string
	TaskType     string
	ContactName  string
	Request      string
	Response     string
	DraftMessage string
	Timestamp    time.Time
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

// companyColumns is the shared SELECT column list for company_workflows queries.
const companyColumns = `id, company_name, slug, public_id, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, updated_at, agent_task_in_progress, meeting_notes`

func scanCompanyRow(row pgx.Row) (*CompanyRow, error) {
	var c CompanyRow
	err := row.Scan(&c.ID, &c.CompanyName, &c.Slug, &c.PublicID, &c.StartedAt, &c.Status,
		&c.ElapsedDays, &c.OutreachCount, &c.ContactCount, &c.RestartCount,
		&c.CurrentContactRole, &c.MeetingBookedAt, &c.LastSnapshotAt, &c.UpdatedAt,
		&c.AgentTaskInProgress, &c.MeetingNotes)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpsertWorkflow inserts or updates the cached workflow state.
// Called by the snapshot activity whenever workflow state changes.
func (r *CompanyRepository) UpsertWorkflow(ctx context.Context, row *CompanyRow) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO company_workflows (id, company_name, slug, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, agent_task_in_progress, meeting_notes, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			elapsed_days = EXCLUDED.elapsed_days,
			outreach_count = EXCLUDED.outreach_count,
			contact_count = EXCLUDED.contact_count,
			restart_count = EXCLUDED.restart_count,
			current_contact_role = EXCLUDED.current_contact_role,
			meeting_booked_at = EXCLUDED.meeting_booked_at,
			last_snapshot_at = EXCLUDED.last_snapshot_at,
			agent_task_in_progress = EXCLUDED.agent_task_in_progress,
			meeting_notes = EXCLUDED.meeting_notes,
			updated_at = NOW()
	`, row.ID, row.CompanyName, row.Slug, row.StartedAt, row.Status,
		row.ElapsedDays, row.OutreachCount, row.ContactCount, row.RestartCount,
		row.CurrentContactRole, row.MeetingBookedAt, row.LastSnapshotAt,
		row.AgentTaskInProgress, row.MeetingNotes)
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

// PersistFullState upserts workflow state, replaces contacts and outreach
// attempts, and optionally inserts an activity feed row — all in one transaction.
func (r *CompanyRepository) PersistFullState(ctx context.Context, row *CompanyRow, contacts []ContactRow, attempts []OutreachAttemptRow, event *ActivityFeedRow) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert company row.
	_, err = tx.Exec(ctx, `
		INSERT INTO company_workflows (id, company_name, slug, started_at, status, elapsed_days, outreach_count, contact_count, restart_count, current_contact_role, meeting_booked_at, last_snapshot_at, agent_task_in_progress, meeting_notes, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			elapsed_days = EXCLUDED.elapsed_days,
			outreach_count = EXCLUDED.outreach_count,
			contact_count = EXCLUDED.contact_count,
			restart_count = EXCLUDED.restart_count,
			current_contact_role = EXCLUDED.current_contact_role,
			meeting_booked_at = EXCLUDED.meeting_booked_at,
			last_snapshot_at = EXCLUDED.last_snapshot_at,
			agent_task_in_progress = EXCLUDED.agent_task_in_progress,
			meeting_notes = EXCLUDED.meeting_notes,
			updated_at = NOW()
	`, row.ID, row.CompanyName, row.Slug, row.StartedAt, row.Status,
		row.ElapsedDays, row.OutreachCount, row.ContactCount, row.RestartCount,
		row.CurrentContactRole, row.MeetingBookedAt, row.LastSnapshotAt,
		row.AgentTaskInProgress, row.MeetingNotes)
	if err != nil {
		return fmt.Errorf("upsert workflow: %w", err)
	}

	// Replace contacts.
	if _, err = tx.Exec(ctx, `DELETE FROM workflow_contacts WHERE workflow_id = $1`, row.ID); err != nil {
		return fmt.Errorf("delete contacts: %w", err)
	}
	for _, c := range contacts {
		_, err = tx.Exec(ctx, `
			INSERT INTO workflow_contacts (workflow_id, name, role, linkedin, active, added_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, row.ID, c.Name, c.Role, c.LinkedIn, c.Active, c.AddedAt)
		if err != nil {
			return fmt.Errorf("insert contact: %w", err)
		}
	}

	// Replace outreach attempts.
	if _, err = tx.Exec(ctx, `DELETE FROM workflow_outreach_attempts WHERE workflow_id = $1`, row.ID); err != nil {
		return fmt.Errorf("delete outreach attempts: %w", err)
	}
	for _, a := range attempts {
		_, err = tx.Exec(ctx, `
			INSERT INTO workflow_outreach_attempts (workflow_id, timestamp, channel, notes, contact)
			VALUES ($1, $2, $3, $4, $5)
		`, row.ID, a.Timestamp, a.Channel, a.Notes, a.Contact)
		if err != nil {
			return fmt.Errorf("insert outreach attempt: %w", err)
		}
	}

	// Optional activity feed event.
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

// PersistStateAndEvent is kept for backward compatibility but delegates to PersistFullState
// with empty contacts and attempts slices.
func (r *CompanyRepository) PersistStateAndEvent(ctx context.Context, row *CompanyRow, event *ActivityFeedRow) error {
	return r.PersistFullState(ctx, row, nil, nil, event)
}

// ListCompanies returns all tracked companies (for the public dashboard).
func (r *CompanyRepository) ListCompanies(ctx context.Context) ([]CompanyRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+companyColumns+`
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
			&c.CurrentContactRole, &c.MeetingBookedAt, &c.LastSnapshotAt, &c.UpdatedAt,
			&c.AgentTaskInProgress, &c.MeetingNotes); err != nil {
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
	row := r.pool.QueryRow(ctx, `
		SELECT `+companyColumns+`
		FROM company_workflows
		WHERE slug = $1
	`, slug)
	return scanCompanyRow(row)
}

// GetCompanyByPublicID returns a single company by its public-facing UUID.
func (r *CompanyRepository) GetCompanyByPublicID(ctx context.Context, publicID string) (*CompanyRow, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+companyColumns+`
		FROM company_workflows
		WHERE public_id = $1
	`, publicID)
	return scanCompanyRow(row)
}

// GetContacts returns all contacts for a workflow, ordered by added_at.
func (r *CompanyRepository) GetContacts(ctx context.Context, workflowID string) ([]ContactRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workflow_id, name, role, linkedin, active, added_at
		FROM workflow_contacts
		WHERE workflow_id = $1
		ORDER BY added_at ASC
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get contacts: %w", err)
	}
	defer rows.Close()

	var contacts []ContactRow
	for rows.Next() {
		var c ContactRow
		if err := rows.Scan(&c.ID, &c.WorkflowID, &c.Name, &c.Role, &c.LinkedIn, &c.Active, &c.AddedAt); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}

// GetOutreachAttempts returns all outreach attempts for a workflow, ordered by timestamp.
func (r *CompanyRepository) GetOutreachAttempts(ctx context.Context, workflowID string) ([]OutreachAttemptRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workflow_id, timestamp, channel, notes, contact
		FROM workflow_outreach_attempts
		WHERE workflow_id = $1
		ORDER BY timestamp ASC
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get outreach attempts: %w", err)
	}
	defer rows.Close()

	var attempts []OutreachAttemptRow
	for rows.Next() {
		var a OutreachAttemptRow
		if err := rows.Scan(&a.ID, &a.WorkflowID, &a.Timestamp, &a.Channel, &a.Notes, &a.Contact); err != nil {
			return nil, fmt.Errorf("scan outreach attempt: %w", err)
		}
		attempts = append(attempts, a)
	}
	return attempts, nil
}

// GetAgentSuggestions returns all agent suggestions for a workflow, ordered by timestamp.
func (r *CompanyRepository) GetAgentSuggestions(ctx context.Context, workflowID string) ([]AgentSuggestionRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workflow_id, task_type, COALESCE(contact_name, ''), COALESCE(request, ''), COALESCE(response, ''), COALESCE(draft_message, ''), COALESCE(timestamp, created_at)
		FROM agent_suggestions
		WHERE workflow_id = $1
		ORDER BY COALESCE(timestamp, created_at) ASC
	`, workflowID)
	if err != nil {
		return nil, fmt.Errorf("get agent suggestions: %w", err)
	}
	defer rows.Close()

	var suggestions []AgentSuggestionRow
	for rows.Next() {
		var s AgentSuggestionRow
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.TaskType, &s.ContactName, &s.Request, &s.Response, &s.DraftMessage, &s.Timestamp); err != nil {
			return nil, fmt.Errorf("scan agent suggestion: %w", err)
		}
		suggestions = append(suggestions, s)
	}
	return suggestions, nil
}

// InsertAgentSuggestion persists an agent suggestion to the database.
func (r *CompanyRepository) InsertAgentSuggestion(ctx context.Context, workflowID, taskType, contactName, request, response, draftMessage string, timestamp time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO agent_suggestions (workflow_id, task_type, contact_name, request, response, draft_message, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, workflowID, taskType, contactName, request, response, draftMessage, timestamp)
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
