package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository is the interface the service depends on.
type Repository interface {
	// CreateUser inserts a new user row. Returns apperr.Conflict on duplicate email.
	CreateUser(ctx context.Context, tx pgx.Tx, u *User) (uuid.UUID, error)

	// GetByID returns a user by their UUID.
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail returns a user by email.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// createEmaiilVerification creates a new email verification record for the user.
	CreateEmailVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, token string, expiresAt time.Time) error

	// GetVerificationByToken retrieves an email verification record by its token.
	GetVerificationByToken(ctx context.Context, token string) (*EmailVerification, error)

	// MarkEmailVerified marks the user's email as verified and sets the verification timestamp.
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error

	//Get User If by Email
	GetUserIdByEmail(ctx context.Context, email string) (uuid.UUID, error)

	// chack emaail exists or not
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	//
	CreateSession(ctx context.Context, s *Session) error

	//
	RevokeSession(ctx context.Context, sessionID uuid.UUID, now time.Time) error

	//
	RevokeAll(ctx context.Context, userID uuid.UUID) error

	//
	GetSessionByRefreshHash(ctx context.Context, hash string) (*Session, error)
}
