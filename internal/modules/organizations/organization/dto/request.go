package dto

type CreateOrganizationRequest struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
}

