package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Env         string `env:"APP_ENV" envDefault:"dev"`
	HTTPAddr    string `env:"HTTP_ADDR" envDefault:":8080"`
	APIKey      string `env:"API_KEY" envDefault:"apitest"`
	DatabaseURL string `env:"DATABASE_URL"`
}

// Load reads environment variables (optionally from .env) into Config.
func Load() Config {
	_ = godotenv.Load() // best-effort .env loader
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	// Default DATABASE_URL based on env if not provided
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "postgres://user:pass@localhost:5432/kart?sslmode=disable"
	}
	return cfg
}
