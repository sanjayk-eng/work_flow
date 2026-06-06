package password

import "golang.org/x/crypto/argon2"

// pure function: no encoding, no IO, no business logic
func generateHash(password string, salt []byte, cfg config) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		cfg.Iterations,
		cfg.Memory,
		cfg.Parallelism,
		cfg.KeyLength,
	)
}
