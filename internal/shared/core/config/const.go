package config

const (
	// App env keys
	EnvAppPort = "APP_PORT"
	EnvAppEnv  = "APP_ENV"

	// DB env keys
	EnvDBHost  = "DB_HOST"
	EnvDBPort  = "DB_PORT"
	EnvDBUser  = "DB_USER"
	EnvDBPass  = "DB_PASS"
	EnvDBName  = "DB_NAME"
	EnvSSLMode = "SSL_MODE"

	// Redis env keys
	EnvRedisHost = "REDIS_HOST"
	EnvRedisPort = "REDIS_PORT"

	// JWT env keys
	EnvJWTSecret        = "JWT_SECRET"
	EnvJWTAccessExpiry  = "JWT_ACCESS_EXPIRY"
	EnvJWTRefreshExpiry = "JWT_REFRESH_EXPIRY"
	EnvJWTIssuer        = "JWT_ISSUER"

	// Log env keys
	EnvLogLevel = "LOG_LEVEL"
)
