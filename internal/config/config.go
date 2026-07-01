package config

import "os"

type Config struct {
	PostgresURL string
	RedisAddr   string
}

func Load() *Config {
	return &Config{
		PostgresURL: getEnv(
			"POSTGRES_URL",
			"postgres://admin:password@localhost:5432/ratelimiter?sslmode=disable",
		),
		RedisAddr: getEnv(
			"REDIS_ADDR",
			"localhost:6379",
		),
	}
}

func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}