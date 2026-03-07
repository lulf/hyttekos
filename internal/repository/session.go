package repository

import (
	"database/sql"
	"fmt"
	"time"

	"hyttekos/internal/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Store saves a new session
func (r *SessionRepository) Store(session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) 
		DO UPDATE SET 
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at
	`
	
	_, err := r.db.Exec(query,
		session.ID,
		session.UserID,
		session.AccessToken,
		session.RefreshToken,
		session.ExpiresAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}
	
	return nil
}

// GetByID retrieves a session by ID
func (r *SessionRepository) GetByID(id string) (*models.Session, error) {
	query := `
		SELECT id, user_id, access_token, refresh_token, expires_at, created_at
		FROM sessions
		WHERE id = $1
	`
	
	var session models.Session
	err := r.db.QueryRow(query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessToken,
		&session.RefreshToken,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Session not found or expired
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return &session, nil
}

// Delete removes a session
func (r *SessionRepository) Delete(id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	return nil
}

// CleanupExpired removes all expired sessions
func (r *SessionRepository) CleanupExpired() (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < $1`

	result, err := r.db.Exec(query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get cleanup count: %w", err)
	}
	
	return count, nil
}

// UpdateTokens updates the tokens for an existing session
func (r *SessionRepository) UpdateTokens(id, accessToken, refreshToken string, expiresAt time.Time) error {
	query := `
		UPDATE sessions
		SET access_token = $1, refresh_token = $2, expires_at = $3
		WHERE id = $4
	`

	result, err := r.db.Exec(query, accessToken, refreshToken, expiresAt, id)
	if err != nil {
		return fmt.Errorf("failed to update session tokens: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("no session found with id %s", id)
	}
	
	return nil
}