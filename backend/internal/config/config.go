package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (Config, error) {
	loadDotEnv()
	cfg := Config{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.Port == "" {
		return Config{}, errors.New("PORT IS REQUIRED")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL IS REQUIRED")
	}
	return cfg, nil
}

func loadDotEnv() {
	paths := []string{filepath.Join("backend", ".env"), ".env"}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			line = strings.TrimPrefix(line, "export ")
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			key, value = strings.TrimSpace(key), strings.TrimSpace(value)
			value = strings.Trim(value, "\"'")
			if key != "" {
				if _, exists := os.LookupEnv(key); !exists {
					_ = os.Setenv(key, value)
				}
			}
		}
		return
	}
}
