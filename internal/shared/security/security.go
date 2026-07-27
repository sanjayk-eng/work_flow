package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func GenerateRefreshToken() (rawToken string, hashedToken string, err error) {

	// 1. create secure random bytes
	b := make([]byte, 32)

	_, err = rand.Read(b)
	if err != nil {
		return "", "", err
	}

	// 2. convert to URL-safe string (send to client)
	rawToken = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)

	// 3. hash token (store in DB)
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	return rawToken, hashedToken, nil
}
