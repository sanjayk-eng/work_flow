package auth

import (
	"context"
	"errors"

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

func (r *postgresRepository) WithTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return r.db.WithTx(ctx, fn)
}

func (r *postgresRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email,
	).Scan(&exists)
	if err != nil {
		return false, apperr.Internal("failed to check email", err)
	}
	return exists, nil
}

func (r *postgresRepository) CreateUser(ctx context.Context, tx pgx.Tx, u *User) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, email_verified, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		u.ID, u.Email, u.PasswordHash, u.EmailVerified, u.Status, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return apperr.Internal("failed to create user", err)
	}
	return nil
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

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, email_verified, status, created_at, updated_at
		FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound("user not found")
		}
		return nil, apperr.Internal("failed to get user", err)
	}
	return u, nil
}

func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, email_verified, status, created_at, updated_at
		FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NotFound("user not found")
		}
		return nil, apperr.Internal("failed to get user", err)
	}
	return u, nil
}
