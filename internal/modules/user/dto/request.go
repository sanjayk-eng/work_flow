package dto

type CreateProfileRequest struct {
	UserID    string `json:"user_id"    validate:"required,uuid4"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name"  validate:"required,min=2,max=100"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url"`
}
