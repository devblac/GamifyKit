package config

import (
	"fmt"
	"os"
	"time"

	"gamifykit/adapters/redis"
	"gamifykit/adapters/sqlx"
)

// ProfileDevelopment returns a configuration optimized for local development
func ProfileDevelopment() *Config {
	cfg := DefaultConfig()
	cfg.Environment = EnvDevelopment
	cfg.Profile = "development"

	// Development-specific settings
	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "text"
	cfg.Server.Address = ":8080"

	// Use in-memory storage for development
	cfg.Storage.Adapter = "memory"

	// Enable metrics but with local defaults
	cfg.Metrics.Enabled = true
	cfg.Metrics.Address = ":9090"

	return cfg
}

// ProfileTesting returns a configuration optimized for automated testing
func ProfileTesting() *Config {
	cfg := DefaultConfig()
	cfg.Environment = EnvTesting
	cfg.Profile = "testing"

	// Testing-specific settings
	cfg.Logging.Level = "warn"
	cfg.Logging.Format = "json"
	cfg.Logging.Output = "stderr"
	cfg.Server.Address = ":0" // Random available port

	// Use in-memory storage for testing
	cfg.Storage.Adapter = "memory"

	// Disable metrics during testing
	cfg.Metrics.Enabled = false

	// Disable rate limiting for tests
	cfg.Security.EnableRateLimit = false

	return cfg
}

// ProfileStaging returns a configuration for staging/pre-production environments
func ProfileStaging() *Config {
	cfg := DefaultConfig()
	cfg.Environment = EnvStaging
	cfg.Profile = "staging"

	// Staging-specific settings
	cfg.Logging.Level = "info"
	cfg.Logging.Format = "json"
	cfg.Server.Address = ":8080"

	// Use file storage for staging (persistent but simple)
	cfg.Storage.Adapter = "file"
	cfg.Storage.File.Path = "/data/gamifykit-staging.json"

	// Enable metrics
	cfg.Metrics.Enabled = true
	cfg.Metrics.Address = ":9090"

	// Enable basic security features
	cfg.Security.EnableRateLimit = true
	cfg.Security.RateLimit.RequestsPerMinute = 120
	cfg.Security.RateLimit.BurstSize = 20

	return cfg
}

// ProfileProduction returns a configuration optimized for production deployments
func ProfileProduction() *Config {
	cfg := DefaultConfig()
	cfg.Environment = EnvProduction
	cfg.Profile = "production"

	// Production-specific settings
	cfg.Logging.Level = "info"
	cfg.Logging.Format = "json"
	cfg.Server.Address = ":8080"

	// Use Redis for production storage
	cfg.Storage.Adapter = "redis"
	cfg.Storage.Redis = redis.Config{
		Addr:         getEnvOrDefault("REDIS_ADDR", "redis:6379"),
		Password:     getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// Enable comprehensive metrics
	cfg.Metrics.Enabled = true
	cfg.Metrics.Address = ":9090"
	cfg.Metrics.CollectSystem = true

	// Enable security features
	cfg.Security.EnableRateLimit = true
	cfg.Security.RateLimit.RequestsPerMinute = 300
	cfg.Security.RateLimit.BurstSize = 50

	// Production server timeouts
	cfg.Server.ReadTimeout = 15 * time.Second
	cfg.Server.WriteTimeout = 15 * time.Second
	cfg.Server.IdleTimeout = 120 * time.Second
	cfg.Server.ReadHeaderTimeout = 10 * time.Second
	cfg.Server.ShutdownTimeout = 60 * time.Second

	return cfg
}

// ProfileProductionSQL returns a production configuration using SQL storage
func ProfileProductionSQL() *Config {
	cfg := ProfileProduction()
	cfg.Profile = "production-sql"

	// Use PostgreSQL for production storage
	cfg.Storage.Adapter = "sql"
	cfg.Storage.SQL = sqlx.Config{
		Driver:          sqlx.DriverPostgres,
		DSN:             getEnvOrDefault("DATABASE_URL", "postgres://gamifykit:gamifykit@postgres:5432/gamifykit?sslmode=require"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	return cfg
}

// ProfileProductionMySQL returns a production configuration using MySQL storage
func ProfileProductionMySQL() *Config {
	cfg := ProfileProduction()
	cfg.Profile = "production-mysql"

	// Use MySQL for production storage
	cfg.Storage.Adapter = "sql"
	cfg.Storage.SQL = sqlx.Config{
		Driver:          sqlx.DriverMySQL,
		DSN:             getEnvOrDefault("DATABASE_URL", "gamifykit:gamifykit@tcp(mysql:3306)/gamifykit?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	return cfg
}

// LoadProfile loads a configuration profile by name
func LoadProfile(profileName string) (*Config, error) {
	switch profileName {
	case "development", "dev":
		return ProfileDevelopment(), nil
	case "testing", "test":
		return ProfileTesting(), nil
	case "staging", "stage":
		return ProfileStaging(), nil
	case "production", "prod":
		return ProfileProduction(), nil
	case "production-sql", "prod-sql":
		return ProfileProductionSQL(), nil
	case "production-mysql", "prod-mysql":
		return ProfileProductionMySQL(), nil
	default:
		return nil, fmt.Errorf("unknown profile: %s", profileName)
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
