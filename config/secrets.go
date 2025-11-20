package config

import (
	"context"
	"fmt"
	"os"
)

// SecretStore defines the interface for external secret management
type SecretStore interface {
	// Get retrieves a secret value by key
	Get(ctx context.Context, key string) (string, error)

	// GetWithDefault retrieves a secret value by key, returning defaultValue if not found
	GetWithDefault(ctx context.Context, key, defaultValue string) string
}

// EnvironmentSecretStore implements SecretStore using environment variables
type EnvironmentSecretStore struct{}

// NewEnvironmentSecretStore creates a new environment-based secret store
func NewEnvironmentSecretStore() *EnvironmentSecretStore {
	return &EnvironmentSecretStore{}
}

// Get retrieves a secret from environment variables
func (e *EnvironmentSecretStore) Get(ctx context.Context, key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("secret %s not found in environment", key)
	}
	return value, nil
}

// GetWithDefault retrieves a secret from environment variables with a default
func (e *EnvironmentSecretStore) GetWithDefault(ctx context.Context, key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LoadSecrets loads sensitive configuration values from a secret store
func (c *Config) LoadSecrets(ctx context.Context, store SecretStore) error {
	// Load database credentials
	if c.Storage.Adapter == "sql" {
		if dsn, err := store.Get(ctx, "GAMIFYKIT_DATABASE_DSN"); err == nil {
			c.Storage.SQL.DSN = dsn
		} else if c.Environment == EnvProduction {
			return fmt.Errorf("database DSN secret required in production: %w", err)
		}
	}

	// Load Redis credentials
	if c.Storage.Adapter == "redis" {
		if password, err := store.Get(ctx, "GAMIFYKIT_REDIS_PASSWORD"); err == nil {
			c.Storage.Redis.Password = password
		}
	}

	// Load any additional secrets that might be needed
	// This is extensible for future secret requirements

	return nil
}

// LoadSecretsFromEnv loads secrets from environment variables (convenience method)
func (c *Config) LoadSecretsFromEnv(ctx context.Context) error {
	store := NewEnvironmentSecretStore()
	return c.LoadSecrets(ctx, store)
}

// ValidateSecrets validates that required secrets are available
func (c *Config) ValidateSecrets(ctx context.Context, store SecretStore) error {
	var errs []string

	// Check required secrets based on configuration
	if c.Storage.Adapter == "sql" && c.Environment == EnvProduction {
		if _, err := store.Get(ctx, "GAMIFYKIT_DATABASE_DSN"); err != nil {
			errs = append(errs, "database DSN secret is required in production")
		}
	}

	// Add more secret validation as needed

	if len(errs) > 0 {
		return fmt.Errorf("secret validation failed: %v", errs)
	}

	return nil
}

// RedactSecrets returns a copy of the config with sensitive values redacted
func (c *Config) RedactSecrets() *Config {
	cfg := *c // Shallow copy

	// Redact database DSN
	if cfg.Storage.SQL.DSN != "" {
		cfg.Storage.SQL.DSN = "[REDACTED]"
	}

	// Redact Redis password
	if cfg.Storage.Redis.Password != "" {
		cfg.Storage.Redis.Password = "[REDACTED]"
	}

	// Add more redactions as needed for future sensitive fields

	return &cfg
}
