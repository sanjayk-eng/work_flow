package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// Session represents a row in the sessions table.
type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

// Repository is the contract the auth service depends on.
// Implemented by the postgres session repository.
type Repository interface {
	// Create inserts a new session row.
	Create(ctx context.Context, s *Session) error

	// GetByTokenHash looks up a session by the hashed refresh token.
	GetByTokenHash(ctx context.Context, tokenHash string) (*Session, error)

	// Revoke marks a single session as revoked.
	Revoke(ctx context.Context, sessionID string) error

	// RevokeAll marks ALL sessions for a user as revoked.
	// Called on reuse detection — forces full re-authentication on all devices.
	RevokeAll(ctx context.Context, userID string) error
}

// GenerateRefreshToken generates a cryptographically random 32-byte token.
// Returns the raw token (sent to client) and its SHA-256 hex hash (stored in DB).
func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = base64.URLEncoding.EncodeToString(b)
	hash = hashToken(raw)
	return
}

// HashToken hashes a raw refresh token for DB lookup.
func HashToken(raw string) string {
	return hashToken(raw)
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func generateID() string {
	return uuid.New().String()
}
