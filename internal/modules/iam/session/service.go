package session

import (
	"context"
	"errors"
	"time"

	"github/sanjay-khandelwal/internal/shared/apperr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service encapsulates all session lifecycle operations.
type Service interface {
	// Create persists a new session.
	Create(ctx context.Context, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) error

	// GetByRefreshHash looks up a session by hashed refresh token.
	GetByRefreshHash(ctx context.Context, hash string) (*Session, error)

	// Revoke invalidates a single session by ID.
	Revoke(ctx context.Context, sessionID uuid.UUID) error

	// RevokeAll invalidates all sessions for a user (refresh token reuse attack).
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, refreshTokenHash string, expiresAt time.Time) error {
	sess := &Session{
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt,
	}
	if err := s.repo.Create(ctx, sess); err != nil {
		return apperr.Internal("failed to create session", err)
	}
	return nil
}

func (s *service) GetByRefreshHash(ctx context.Context, hash string) (*Session, error) {
	sess, err := s.repo.GetByRefreshHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.Unauthorized("invalid refresh token")
		}
		return nil, apperr.Internal("failed to get session", err)
	}
	return sess, nil
}

func (s *service) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.repo.Revoke(ctx, sessionID); err != nil {
		return apperr.Internal("failed to revoke session", err)
	}
	return nil
}

func (s *service) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.RevokeAll(ctx, userID); err != nil {
		return apperr.Internal("failed to revoke all sessions", err)
	}
	return nil
}
