package organization

import (
	"context"

	"github/sanjay-khandelwal/internal/shared/core/database/postgres"

	"github.com/google/uuid"
)

type postgresRepository struct {
	db *postgres.DB
}

func NewRepository(db *postgres.DB) Repository {
	return &postgresRepository{db: db}
}

// CREATE
func (r *postgresRepository) CreateOrganization(ctx context.Context, org *Organization) error {

	query := `
		INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		org.ID,
		org.Name,
		org.Slug,
		org.OwnerID,
		org.CreatedAt,
		org.UpdatedAt,
	)

	return err
}

// GET BY ID
func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {

	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at, deleted_at
		FROM organizations
		WHERE id = $1 AND deleted_at IS NULL
	`

	var org Organization

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.DeletedAt,
	)

	return &org, err
}

// GET BY SLUG
func (r *postgresRepository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {

	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at, deleted_at
		FROM organizations
		WHERE slug = $1 AND deleted_at IS NULL
	`

	var org Organization

	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.DeletedAt,
	)

	return &org, err
}

// LIST BY OWNER (simple MVP)
func (r *postgresRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Organization, error) {

	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at, deleted_at
		FROM organizations
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*Organization

	for rows.Next() {
		var org Organization

		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.OwnerID,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		orgs = append(orgs, &org)
	}

	return orgs, nil
}

// UPDATE
func (r *postgresRepository) UpdateOrganization(ctx context.Context, org *Organization) error {

	query := `
		UPDATE organizations
		SET name = $1,
		    slug = $2,
		    updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`

	_, err := r.db.Pool.Exec(ctx, query,
		org.Name,
		org.Slug,
		org.ID,
	)

	return err
}

// SOFT DELETE
func (r *postgresRepository) DeleteOrganization(ctx context.Context, id uuid.UUID) error {

	query := `
		UPDATE organizations
		SET deleted_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// EXISTS BY SLUG
func (r *postgresRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {

	query := `
		SELECT EXISTS(
			SELECT 1 FROM organizations
			WHERE slug = $1 AND deleted_at IS NULL
		)
	`

	var exists bool

	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
