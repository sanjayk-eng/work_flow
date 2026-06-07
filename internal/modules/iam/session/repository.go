package session

import (
	"context"

	"github.com/google/uuid"
)

// Repository handles all session persistence.
type Repository interface {
	// Create inserts a new session row.
	Create(ctx context.Context, s *Session) error

	// GetByRefreshHash retrieves a session by its hashed refresh token.
	GetByRefreshHash(ctx context.Context, hash string) (*Session, error)

	// Revoke marks a single session as revoked.
	Revoke(ctx context.Context, sessionID uuid.UUID) error

	// RevokeAll revokes every active session for a user (reuse-attack protection).
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}
