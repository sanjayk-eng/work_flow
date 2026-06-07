package auth

import (
	"context"
	"time"

	"github/sanjay-khandelwal/internal/shared/core/database/postgres"

	"github.com/google/uuid"
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

func (r *postgresRepository) CreateUser(ctx context.Context, tx pgx.Tx, u *User) (uuid.UUID, error) {

	const query = `
		INSERT INTO users (
			email,
			password_hash,
			email_verified,
			status,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id uuid.UUID

	err := tx.QueryRow(
		ctx,
		query,
		u.Email,
		u.PasswordHash,
		u.EmailVerified,
		u.Status,
		u.CreatedAt,
		u.UpdatedAt,
	).Scan(&id)

	return id, err
}
func (r *postgresRepository) GetByID(ctx context.Context, id string) (*User, error) {

	var u User

	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, email_verified, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.EmailVerified,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
func (r *postgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {

	var u User
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, email_verified, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.EmailVerified,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
func (r *postgresRepository) CreateEmailVerification(ctx context.Context, tx pgx.Tx, userID uuid.UUID, token string, expiresAt time.Time) error {

	_, err := tx.Exec(ctx, `
		INSERT INTO email_verifications (
			id,
			user_id,
			token,
			expires_at,
			created_at
		)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
	`,
		userID,
		token,
		expiresAt,
	)

	return err
}
func (r *postgresRepository) GetVerificationByToken(ctx context.Context, token string) (*EmailVerification, error) {

	var ev EmailVerification

	err := r.db.Pool.QueryRow(ctx, `
		SELECT
			id,
			user_id,
			token,
			expires_at,
			verified_at,
			created_at
		FROM email_verifications
		WHERE token = $1
	`, token).Scan(
		&ev.ID,
		&ev.UserID,
		&ev.Token,
		&ev.ExpiresAt,
		&ev.VerifiedAt,
		&ev.CreatedAt,
	)

	return &ev, err
}

func (r *postgresRepository) MarkEmailVerified(
	ctx context.Context,
	userID uuid.UUID,
) error {

	cmd, err := r.db.Pool.Exec(ctx, `UPDATE users SET email_verified = TRUE WHERE id = $1  AND email_verified = FALSE`, userID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	_, err = r.db.Pool.Exec(ctx, `UPDATE email_verifications SET verified_at = NOW() WHERE user_id = $1  AND verified_at IS NULL`, userID)
	if err != nil {
		return err
	}

	return nil
}
func (r *postgresRepository) GetUserIdByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	var UserId uuid.UUID
	err := r.db.Pool.QueryRow(ctx, `SELECT idFROM usersWHERE email = $1`, email).Scan(UserId)

	return UserId, err
}
