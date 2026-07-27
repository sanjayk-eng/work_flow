package jwt

import "time"

type config struct {
	SecretKey       []byte // always []byte — no string conversion risk
	AccessExpiry    time.Duration
	Issuer          string
	AccessTokenTTL  time.Time
	RefreshTokenTTL time.Time
}

func defaultConfig() *config {
	return &config{
		SecretKey:       []byte("change-this-secret"),
		Issuer:          "app-service",
		AccessTokenTTL:  time.Now().Add(15 * time.Minute),
		RefreshTokenTTL: time.Now().Add(7 * 24 * time.Hour),
	}
}
