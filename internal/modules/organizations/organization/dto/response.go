package dto

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationResponse returned to API clients
type OrganizationResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	OwnerID uuid.UUID `json:"owner_id"`

	CreatedAt time.Time `json:"created_at"`
}
