package auth

import "time"

type UserStatus string
type email string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
)

// User maps to the users table.
type User struct {
	ID            string
	Email         string
	PasswordHash  string
	EmailVerified bool
	Status        UserStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type EmailVerification struct {
	ID         string
	UserID     string
	Token      string
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	CreatedAt  time.Time
}
