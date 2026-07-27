package organization

import (
	"context"
	"github/sanjay-khandelwal/internal/modules/organizations/organization/dto"
	"github/sanjay-khandelwal/internal/shared/apperr"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, ownerID uuid.UUID, req *dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.OrganizationResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.OrganizationResponse, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*dto.OrganizationResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, ownerID uuid.UUID, req *dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error) {

	// 1. check slug uniqueness
	slug := generateSlug(req.Name)

	exists, err := s.repo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, apperr.Internal("failed to check slug", err)
	}
	if exists {
		return nil, apperr.Conflict("organization already exists")
	}

	org := &Organization{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      slug,
		OwnerID:   ownerID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateOrganization(ctx, org); err != nil {
		return nil, apperr.Internal("failed to create organization", err)
	}

	return toResponse(org), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*dto.OrganizationResponse, error) {

	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal("failed to get organization", err)
	}

	return toResponse(org), nil
}

func (s *service) GetBySlug(ctx context.Context, slug string) (*dto.OrganizationResponse, error) {

	org, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperr.Internal("failed to get organization", err)
	}

	return toResponse(org), nil
}
func (s *service) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*dto.OrganizationResponse, error) {

	orgs, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, apperr.Internal("failed to list organizations", err)
	}

	res := make([]*dto.OrganizationResponse, 0, len(orgs))

	for _, o := range orgs {
		res = append(res, toResponse(o))
	}

	return res, nil
}
func (s *service) Update(ctx context.Context, id uuid.UUID, req *dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error) {

	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.Internal("organization not found", err)
	}

	org.Name = req.Name
	org.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateOrganization(ctx, org); err != nil {
		return nil, apperr.Internal("failed to update organization", err)
	}

	return toResponse(org), nil
}
func (s *service) Delete(ctx context.Context, id uuid.UUID) error {

	if err := s.repo.DeleteOrganization(ctx, id); err != nil {
		return apperr.Internal("failed to delete organization", err)
	}

	return nil
}

func toResponse(org *Organization) *dto.OrganizationResponse {
	return &dto.OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		OwnerID:   org.OwnerID,
		CreatedAt: org.CreatedAt,
	}
}
func generateSlug(name string) string {

	name = strings.ToLower(strings.TrimSpace(name))

	var builder strings.Builder
	prevDash := false

	for _, r := range name {

		// allow letters & numbers
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
			prevDash = false
			continue
		}

		// convert space or symbol → dash
		if unicode.IsSpace(r) || r == '-' || r == '_' {
			if !prevDash {
				builder.WriteRune('-')
				prevDash = true
			}
		}
	}

	slug := builder.String()

	// trim leading/trailing dash
	slug = strings.Trim(slug, "-")

	return slug
}
