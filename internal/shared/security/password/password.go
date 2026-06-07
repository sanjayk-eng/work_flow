package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// Hasher owns config and exposes secure API
type Hasher struct {
	cfg *config
}

// New creates a Hasher with default Argon2id parameters.
func New() *Hasher {
	return &Hasher{cfg: defaultConfig()}
}

func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.cfg.SaltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := generateHash(password, salt, h.cfg)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.cfg.Memory,
		h.cfg.Iterations,
		h.cfg.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *Hasher) Verify(password, encoded string) (bool, error) {
	var memory uint32
	var iterations uint32
	var parallelism uint8
	var saltB64, hashB64 string

	_, err := fmt.Sscanf(
		encoded,
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		&memory,
		&iterations,
		&parallelism,
		&saltB64,
		&hashB64,
	)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, err
	}

	newHash := generateHash(password, salt, &config{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		KeyLength:   uint32(len(expectedHash)),
	})

	return subtle.ConstantTimeCompare(newHash, expectedHash) == 1, nil
}
