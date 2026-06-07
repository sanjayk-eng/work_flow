package session

import (
	"time"

	"github.com/google/uuid"
)

// Session maps to the sessions table.
// Only the hashed refresh token is stored — never the raw token.
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}
