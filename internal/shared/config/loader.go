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
		},

		Redis: RedisConfig{
			Host: getEnv(EnvRedisHost, "localhost"),
			Port: getEnv(EnvRedisPort, "6379"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
