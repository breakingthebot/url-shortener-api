// src/config/config.go
// Loads runtime configuration for the HTTP server and PostgreSQL connection from environment variables.
// Connects to: src/main.go, .env.example
// Created: 2026-06-28

package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultAppEnv          = "development"
	defaultAppHost         = "127.0.0.1"
	defaultAppPort         = "8080"
	defaultShortCodeLength = 6
)

// Config stores the environment-driven settings for the application.
type Config struct {
	AppEnv          string
	AppHost         string
	AppPort         string
	DatabaseURL     string
	ShortCodeLength int
}

// LoadConfig reads configuration from environment variables and applies safe defaults.
func LoadConfig() (Config, error) {
	shortCodeLength, err := loadShortCodeLength()
	if err != nil {
		return Config{}, err
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return Config{
		AppEnv:          getEnvOrDefault("APP_ENV", defaultAppEnv),
		AppHost:         getEnvOrDefault("APP_HOST", defaultAppHost),
		AppPort:         getEnvOrDefault("APP_PORT", defaultAppPort),
		DatabaseURL:     databaseURL,
		ShortCodeLength: shortCodeLength,
	}, nil
}

// Address builds the listen address for the HTTP server.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.AppHost, c.AppPort)
}

// loadShortCodeLength validates and returns the configured shortcode length.
func loadShortCodeLength() (int, error) {
	value := getEnvOrDefault("SHORT_CODE_LENGTH", strconv.Itoa(defaultShortCodeLength))
	length, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("SHORT_CODE_LENGTH must be a valid integer: %w", err)
	}

	if length < 4 || length > 12 {
		return 0, fmt.Errorf("SHORT_CODE_LENGTH must be between 4 and 12")
	}

	return length, nil
}

// getEnvOrDefault returns the environment value when set, otherwise the provided default.
func getEnvOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
