package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	RedisAddr      string
	SnapshotMaxAge time.Duration
	Upstream       UpstreamConfig
}

type UpstreamConfig struct {
	SalesURL   string
	ServiceURL string
	Timeout    time.Duration
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	upstreamTimeout, err := durationEnv("UPSTREAM_TIMEOUT", 2*time.Second)
	if err != nil {
		return nil, err
	}
	snapshotMaxAge, err := durationEnv("SNAPSHOT_MAX_AGE", 720*time.Hour)
	if err != nil {
		return nil, err
	}

	return &Config{
		HTTPAddr:       stringEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:    stringEnv("DATABASE_URL", "postgres://app:app@localhost:5432/documents?sslmode=disable"),
		RedisAddr:      stringEnv("REDIS_ADDR", "localhost:6379"),
		SnapshotMaxAge: snapshotMaxAge,
		Upstream: UpstreamConfig{
			SalesURL:   stringEnv("SALES_API_URL", "http://localhost:8081"),
			ServiceURL: stringEnv("SERVICE_API_URL", "http://localhost:8082"),
			Timeout:    upstreamTimeout,
		},
	}, nil
}

// stringEnv returns the environment variable key, or fallback to defaultValue
func stringEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// durationEnv parses the environment variable key as a time.Duration
// (e.g. "2s", "720h"), or fallback to def
func durationEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s=%q is not a valid duration: %w", key, v, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("config: %s must be positive, got %s", key, d)
	}
	return d, nil
}
