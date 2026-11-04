package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	RedisAddr      string
	SnapshotMaxAge time.Duration
	Upstream       UpstreamConfig
	LogLevel       slog.Level
}

type UpstreamConfig struct {
	SalesURL   string
	ServiceURL string
	Timeout    time.Duration // total budget per source, including retries
	Retries    int           // extra attempts after the first one; 0 = no retry
	RetryDelay time.Duration // wait between attempts
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	upstreamTimeout, err := durationEnv("UPSTREAM_TIMEOUT", 2*time.Second)
	if err != nil {
		return nil, err
	}
	upstreamRetries, err := intEnv("UPSTREAM_RETRIES", 1)
	if err != nil {
		return nil, err
	}
	upstreamRetryDelay, err := durationEnv("UPSTREAM_RETRY_DELAY", 200*time.Millisecond)
	if err != nil {
		return nil, err
	}
	snapshotMaxAge, err := durationEnv("SNAPSHOT_MAX_AGE", 720*time.Hour)
	if err != nil {
		return nil, err
	}
	logLevel, err := logLevelEnv("LOG_LEVEL", slog.LevelInfo)
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
			Retries:    upstreamRetries,
			RetryDelay: upstreamRetryDelay,
		},
		LogLevel: logLevel,
	}, nil
}

func logLevelEnv(key string, defaultValue slog.Level) (slog.Level, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue, nil
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(v)); err != nil {
		return 0, fmt.Errorf("config: %s=%q is not a valid log level (debug, info, warn, error): %w", key, v, err)
	}
	return level, nil
}

// stringEnv returns the environment variable key, or fallback to defaultValue
func stringEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// intEnv parses the environment variable key as a non-negative int,
// or fallback to defaultValue
func intEnv(key string, defaultValue int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s=%q is not a valid integer: %w", key, v, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("config: %s must not be negative, got %d", key, n)
	}
	return n, nil
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
