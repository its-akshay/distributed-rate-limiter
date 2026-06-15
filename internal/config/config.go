package config

type Config struct {
	PostgresURL string
	RedisAddr   string
}

func Load() *Config {
	return &Config{
		PostgresURL: "postgres://admin:password@localhost:5432/ratelimiter",
		RedisAddr:   "localhost:6379",
	}
}