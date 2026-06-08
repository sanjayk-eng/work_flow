package organization

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines organization DB operations
type Repository interface {

	// Create organization
	CreateOrganization(ctx context.Context, org *Organization) error

	// Get organization by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)

	// Get organization by slug (for URL access)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)

	// List organizations by owner (simple MVP access control)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Organization, error)

	// Update organization
	UpdateOrganization(ctx context.Context, org *Organization) error

	// Soft delete organization
	DeleteOrganization(ctx context.Context, id uuid.UUID) error

	// Check if slug already exists
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
}
