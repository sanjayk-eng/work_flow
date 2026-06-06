package jwt

import (
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Sentinel errors — safe to return to callers, no internal detail leaked.
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Service struct {
	cfg *config
}

func New(secretKey []byte) *Service {
	cfg := defaultConfig()
	if len(secretKey) > 0 {
		cfg.SecretKey = secretKey
	}
	return &Service{cfg: cfg}
}

// CreateToken issues a short-lived access token containing userID and email.
// Refresh token lifecycle is handled externally via DB sessions table.
func (s *Service) CreateToken(userID, email string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.cfg.SecretKey)
}

// VerifyToken parses and validates an access token.
// Returns Claims (userID) on success.
// Internal errors are logged but never exposed to the caller.
func (s *Service) VerifyToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, s.keyFunc)
	if err != nil {
		slog.Debug("jwt: verification failed", "reason", err.Error())
		return nil, s.mapError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		slog.Debug("jwt: claims cast failed or token invalid")
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// keyFunc validates HMAC signing method — rejects algorithm confusion attacks.
func (s *Service) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		slog.Debug("jwt: unexpected signing method", "alg", token.Header["alg"])
		return nil, ErrInvalidToken
	}
	return s.cfg.SecretKey, nil
}

func (s *Service) mapError(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return ErrExpiredToken
	case errors.Is(err, jwt.ErrTokenNotValidYet),
		errors.Is(err, jwt.ErrTokenMalformed),
		errors.Is(err, jwt.ErrTokenSignatureInvalid),
		errors.Is(err, jwt.ErrTokenUnverifiable):
		return ErrInvalidToken
	default:
		return ErrInvalidToken
	}
}
