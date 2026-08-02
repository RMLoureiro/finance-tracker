package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL   string
	Port          string
	SecureCookies bool
	// WebDir aponta para o export estático do Next; vazio em desenvolvimento
	// (o Next dev server faz proxy para esta API).
	WebDir string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          os.Getenv("PORT"),
		SecureCookies: os.Getenv("COOKIE_SECURE") == "true",
		WebDir:        os.Getenv("WEB_DIR"),
	}
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL em falta")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return cfg, nil
}
