package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL   string
	Port          string
	SecureCookies bool
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          os.Getenv("PORT"),
		SecureCookies: os.Getenv("COOKIE_SECURE") == "true",
	}
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL em falta")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg, nil
}
