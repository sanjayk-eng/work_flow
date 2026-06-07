package user

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Repository is the interface the service depends on.
type Repository interface {

	// CreateProfile inserts a new profile row inside the provided transaction.
	CreateProfile(ctx context.Context, tx pgx.Tx, p *Profile) error
}
