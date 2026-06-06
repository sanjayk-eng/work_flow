package session

import (
	"context"
	"log/slog"
	"time"

	"github/sanjay-khandelwal/internal/shared/apperr"
)

// Service handles refresh token lifecycle:
//   - rotation: every use invalidates old token, issues new one
//   - reuse detection: revoked token presented → revoke ALL sessions
type Service struct {
	repo          Repository
	tokenLifetime time.Duration
}

func NewService(repo Repository, tokenLifetime time.Duration) *Service {
	return &Service{repo: repo, tokenLifetime: tokenLifetime}
}

// Issue creates a new session for a user and returns the raw refresh token.
// The raw token is returned to the client; only the hash is stored in DB.
func (s *Service) Issue(ctx context.Context, userID, sessionID string) (string, error) {
	raw, hash, err := GenerateRefreshToken()
	if err != nil {
		return "", apperr.Internal("failed to generate refresh token", err)
	}

	sess := &Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hash,
		ExpiresAt:        time.Now().Add(s.tokenLifetime),
	}

	if err := s.repo.Create(ctx, sess); err != nil {
		return "", apperr.Internal("failed to store session", err)
	}

	return raw, nil
}

// Rotate validates the incoming refresh token, revokes the old session,
// and issues a new session + token. Implements reuse detection.
//
// Flow:
//  1. Hash incoming token → DB lookup
//  2. If not found → invalid token
//  3. If found but revoked → REUSE DETECTED → revoke all sessions → force re-login
//  4. If expired → revoke + return expired error
//  5. Revoke old session → issue new session → return new raw token
func (s *Service) Rotate(ctx context.Context, rawToken, newSessionID string) (string, error) {
	hash := HashToken(rawToken)

	sess, err := s.repo.GetByTokenHash(ctx, hash)
	if err != nil {
		return "", apperr.Unauthorized("invalid refresh token")
	}

	// Reuse detection — token was already revoked
	if sess.RevokedAt != nil {
		slog.Warn("refresh token reuse detected — revoking all sessions",
			"user_id", sess.UserID,
			"session_id", sess.ID,
		)
		_ = s.repo.RevokeAll(ctx, sess.UserID)
		return "", apperr.Unauthorized("token reuse detected — please log in again")
	}

	// Expired
	if time.Now().After(sess.ExpiresAt) {
		_ = s.repo.Revoke(ctx, sess.ID)
		return "", apperr.Unauthorized("refresh token expired")
	}

	// Revoke old session
	if err := s.repo.Revoke(ctx, sess.ID); err != nil {
		return "", apperr.Internal("failed to revoke session", err)
	}

	// Issue new session
	newRaw, err := s.Issue(ctx, sess.UserID, newSessionID)
	if err != nil {
		return "", err
	}

	return newRaw, nil
}

// Revoke explicitly invalidates a session (logout).
func (s *Service) Revoke(ctx context.Context, sessionID string) error {
	if err := s.repo.Revoke(ctx, sessionID); err != nil {
		return apperr.Internal("failed to revoke session", err)
	}
	return nil
}

// RevokeAll invalidates all sessions for a user (logout all devices).
func (s *Service) RevokeAll(ctx context.Context, userID string) error {
	if err := s.repo.RevokeAll(ctx, userID); err != nil {
		return apperr.Internal("failed to revoke all sessions", err)
	}
	return nil
}
