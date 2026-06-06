package dto

import "time"

// UserResponse is returned after successful registration or login.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// UserResponseInput carries the domain data needed to build a UserResponse.
// Service constructs this from User + Profile — dto never imports the auth domain.
type UserResponseInput struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Status    string
	CreatedAt time.Time
}

// NewUserResponse maps a UserResponseInput into a UserResponse.
// All response shaping lives here — service just passes the input struct.
func NewUserResponse(in UserResponseInput) UserResponse {
	return UserResponse{
		ID:        in.ID,
		Email:     in.Email,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Status:    in.Status,
		CreatedAt: in.CreatedAt,
	}
}

// RegisterResponse wraps the created user.
type RegisterResponse struct {
	User UserResponse `json:"user"`
}
