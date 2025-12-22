package config

import (
	"os"
	"time"
)

type Config struct {
	DBUrl     string
	JwtSecret string
	TokenTTL  time.Duration
}

func Load() *Config {
	ttl, _ := time.ParseDuration(os.Getenv("TOKEN_TTL"))

	return &Config{
		DBUrl:     os.Getenv("DB_URL"),
		JwtSecret: os.Getenv("JWT_SECRET"),
		TokenTTL:  ttl,
	}
}
