package session

import (
	"context"
	"time"

	"github/sanjay-khandelwal/internal/shared/core/database/postgres"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *postgres.DB
}

func NewRepository(db *postgres.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, s *Session) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO sessions (user_id, refresh_token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`, s.UserID, s.RefreshTokenHash, s.ExpiresAt)
	return err
}

func (r *postgresRepository) GetByRefreshHash(ctx context.Context, hash string) (*Session, error) {
	var s Session
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked_at, created_at
		FROM sessions
		WHERE refresh_token_hash = $1
	`, hash).Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *postgresRepository) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = $1 WHERE id = $2 AND revoked_at IS NULL
	`, now, sessionID)
	return err
}

func (r *postgresRepository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL
	`, now, userID)
	return err
}
