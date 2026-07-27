package organization

import (
	"time"

	"github.com/google/uuid"
)

// Organization represents a SaaS tenant (company/account root)
type Organization struct {
	ID      uuid.UUID
	Name    string
	Slug    string
	OwnerID uuid.UUID

	DeletedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
