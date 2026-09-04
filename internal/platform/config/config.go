// Package config loads and validates process configuration from the environment
// at startup. A missing or invalid required value is a fatal error.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config is the fully-validated process configuration.
type Config struct {
	DatabaseURL   string
	Port          string
	Env           string
	SessionTTL    time.Duration
	ShutdownGrace time.Duration
}

// IsProduction reports whether the process is running in the production environment.
func (c Config) IsProduction() bool { return c.Env == "production" }

// Load reads configuration from the environment. It returns an error describing the
// first problem it finds; the caller is expected to exit non-zero.
func Load() (Config, error) {
	c := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        envOr("PORT", "8080"),
		Env:         envOr("ENV", "development"),
	}

	if c.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if c.Env != "development" && c.Env != "production" {
		return Config{}, fmt.Errorf("ENV must be \"development\" or \"production\", got %q", c.Env)
	}

	ttl, err := durationOr("SESSION_TTL", 720*time.Hour)
	if err != nil {
		return Config{}, err
	}
	c.SessionTTL = ttl

	grace, err := durationOr("SHUTDOWN_GRACE", 10*time.Second)
	if err != nil {
		return Config{}, err
	}
	c.ShutdownGrace = grace

	return c, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationOr(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s is not a valid duration: %w", key, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s must be positive, got %s", key, d)
	}
	return d, nil
}
