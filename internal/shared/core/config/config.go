package config

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig
	Log   LogConfig
}

type LogConfig struct {
	Level string // debug | info | warn | error
}

type JWTConfig struct {
	SecretKey     string
	AccessExpiry  string // e.g. "15m"
	RefreshExpiry string // e.g. "7d"
	Issuer        string
}

type AppConfig struct {
	Port string
	Env  string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host string
	Port string
}
