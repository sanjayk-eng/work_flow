package postgres

import (
	"fmt"
	"github/sanjay-khandelwal/internal/shared/config"
)

func BuildDSN(cfg config.DBConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)
}
