package jwt

import "time"




type config struct {
	SecretKey    []byte
	AccessExpiry time.Duration
	Issuer       string
}

func defaultConfig() config {
	return config{
		SecretKey:    []byte("change-this-secret"), // replace in prod via env at app layer
		AccessExpiry: 15 * time.Minute,
		Issuer:       "app-service",
	}
}
