package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository handles user and email-verification persistence.
// Session persistence is handled separately by session.Repository.
type Repository interface {
	CreateUser(ctx context.Context, tx pgx.Tx, u *User) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	CreateEmailVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, token string, expiresAt time.Time) error
	GetVerificationByToken(ctx context.Context, token string) (*EmailVerification, error)
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
}
