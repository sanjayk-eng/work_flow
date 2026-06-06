package jwt

import "github.com/golang-jwt/jwt/v5"

// Claims carries identity only — userID + email for convenience.
// No roles, no permissions, no sessionID — refresh lifecycle is managed via DB sessions table.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
