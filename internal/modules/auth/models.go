package auth

import "time"

// User maps to the users table.
type User struct {
	ID            string
	Email         string
	PasswordHash  string
	EmailVerified bool
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Profile maps to the profiles table.
type Profile struct {
	UserID    string
	FirstName string
	LastName  string
	AvatarURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}
