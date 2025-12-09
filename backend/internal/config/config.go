package config

import (
	"os"
)

type Config struct {
	DBUrl     string
	JwtSecret string
}

func Load() *Config {
	return &Config{
		DBUrl:     os.Getenv("DB_URL"),
		JwtSecret: os.Getenv("JWT_SECRET"),
	}
}
