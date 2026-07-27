package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github/sanjay-khandelwal/internal/shared/apperr"
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

// Verify checks a plaintext password against an encoded Argon2id hash.
// Format: $argon2id$v=19$m=<mem>,t=<iter>,p=<par>$<salt>$<hash>
func (h *Hasher) Verify(password, encoded string) (bool, error) {
	// Split on $ — parts: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<hash>"]
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, apperr.Internal("invalid hash format")
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, apperr.Internal("invalid hash params", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, apperr.Internal("invalid hash salt", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, apperr.Internal("invalid hash value", err)
	}

	newHash := generateHash(password, salt, &config{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		KeyLength:   uint32(len(expectedHash)),
	})

	return subtle.ConstantTimeCompare(newHash, expectedHash) == 1, nil
}
