package auth

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Repository is the interface the service depends on.
// The concrete implementation lives in repository_postgres.go.
type Repository interface {
	// ExistsByEmail returns true if a user with that email already exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// CreateUser inserts a new user row inside the provided transaction.
	CreateUser(ctx context.Context, tx pgx.Tx, u *User) error

	// CreateProfile inserts a new profile row inside the provided transaction.
	CreateProfile(ctx context.Context, tx pgx.Tx, p *Profile) error

	// GetByID returns a user by their UUID.
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByEmail returns a user by email.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// WithTx runs fn inside a database transaction.
	// Service calls this instead of holding a *postgres.DB directly.
	WithTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}
