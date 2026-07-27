package config

import "os"

func NewConfig() *Config {
	return &Config{
		App: AppConfig{
			Port: getEnv(EnvAppPort, "8080"),
			Env:  getEnv(EnvAppEnv, "dev"),
		},

		DB: DBConfig{
			Host:     getEnv(EnvDBHost, "localhost"),
			Port:     getEnv(EnvDBPort, "5432"),
			User:     getEnv(EnvDBUser, "postgres"),
			Password: getEnv(EnvDBPass, ""),
			Name:     getEnv(EnvDBName, "app"),
			SSLMode:  getEnv(EnvSSLMode, "disable"),
		},

		Redis: RedisConfig{
			Host: getEnv(EnvRedisHost, "localhost"),
			Port: getEnv(EnvRedisPort, "6379"),
		},

		JWT: JWTConfig{
			SecretKey:     getEnv(EnvJWTSecret, "change-this-secret"),
			AccessExpiry:  getEnv(EnvJWTAccessExpiry, "15m"),
			RefreshExpiry: getEnv(EnvJWTRefreshExpiry, "168h"),
			Issuer:        getEnv(EnvJWTIssuer, "app-service"),
		},
		Log: LogConfig{
			Level: getEnv(EnvLogLevel, "info"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
