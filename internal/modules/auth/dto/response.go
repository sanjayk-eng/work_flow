package dto

import "time"

// UserResponse is returned after successful registration or login.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
