package jwt

import "time"

type config struct {
	SecretKey    []byte        // always []byte — no string conversion risk
	AccessExpiry time.Duration
	Issuer       string
}

func defaultConfig() *config {
	return &config{
		SecretKey:    []byte("change-this-secret"),
		AccessExpiry: 15 * time.Minute,
		Issuer:       "app-service",
	}
}
