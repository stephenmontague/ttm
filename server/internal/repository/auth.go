package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminUser struct {
	ID           int
	Email        string
	PasswordHash string
}

type AdminSession struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
}

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

// UpsertAdminUser inserts the seed user or updates password_hash if email already exists.
func (r *AuthRepository) UpsertAdminUser(ctx context.Context, email, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_users (email, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
	`, email, passwordHash)
	if err != nil {
		return fmt.Errorf("upsert admin user: %w", err)
	}
	return nil
}

// GetAdminUserByEmail returns the admin user or an error if not found.
func (r *AuthRepository) GetAdminUserByEmail(ctx context.Context, email string) (*AdminUser, error) {
	var u AdminUser
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash FROM admin_users WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("get admin user: %w", err)
	}
	return &u, nil
}

// CreateSession generates a random token, inserts a session row, and returns the token.
func (r *AuthRepository) CreateSession(ctx context.Context, userID int, maxAgeSecs int) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(b)

	expiresAt := time.Now().Add(time.Duration(maxAgeSecs) * time.Second)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

// ValidateSession returns the session if the token exists and has not expired.
// On success, it extends the expiry (sliding window) so active sessions stay alive.
// It also lazy-deletes expired sessions for this user in the background.
func (r *AuthRepository) ValidateSession(ctx context.Context, token string, maxAgeSecs int) (*AdminSession, error) {
	var s AdminSession
	err := r.pool.QueryRow(ctx, `
		SELECT token, user_id, expires_at
		FROM admin_sessions
		WHERE token = $1 AND expires_at > NOW()
	`, token).Scan(&s.Token, &s.UserID, &s.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("validate session: %w", err)
	}

	// Sliding expiry: extend the session on each valid request.
	newExpiry := time.Now().Add(time.Duration(maxAgeSecs) * time.Second)
	_, _ = r.pool.Exec(ctx, `
		UPDATE admin_sessions SET expires_at = $1 WHERE token = $2
	`, newExpiry, token)

	// Lazy cleanup: delete expired sessions for this user in the background.
	go func() {
		_, _ = r.pool.Exec(context.Background(), `
			DELETE FROM admin_sessions WHERE user_id = $1 AND expires_at <= NOW()
		`, s.UserID)
	}()

	return &s, nil
}

// DeleteSession removes a specific token (used on logout).
func (r *AuthRepository) DeleteSession(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM admin_sessions WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
