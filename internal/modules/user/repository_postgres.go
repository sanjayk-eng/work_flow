package user

import (
	"context"
	"github/sanjay-khandelwal/internal/shared/apperr"
	"github/sanjay-khandelwal/internal/shared/core/database/postgres"

	"github.com/jackc/pgx/v5"
)

type postgresRepository struct {
	db *postgres.DB
}

func NewRepository(db *postgres.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) CreateProfile(ctx context.Context, tx pgx.Tx, p *Profile) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO profiles (user_id, first_name, last_name, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		p.UserID, p.FirstName, p.LastName, p.AvatarURL, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return apperr.Internal("failed to create profile", err)
	}
	return nil
}
