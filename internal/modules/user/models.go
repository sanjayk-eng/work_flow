package user

import "time"

// Profile maps to the profiles table.
type Profile struct {
	Id        string
	UserID    string
	FirstName string
	LastName  string
	AvatarURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}
