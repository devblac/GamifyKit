# Configuration Management

GamifyKit provides a comprehensive configuration management system that supports environment-based configuration, validation, and secret management.

## Features

- **Environment-based configuration** with environment variable overrides
- **Configuration profiles** for different environments (development, testing, production)
- **Schema validation** with helpful error messages
- **Secret management** integration with external stores
- **JSON file support** for complex configurations

## Quick Start

### Basic Usage

```go
package main

import (
    "gamifykit/config"
)

func main() {
    // Load configuration from environment variables
    cfg, err := config.Load()
    if err != nil {
        panic(err)
    }

    // Use configuration
    fmt.Printf("Server will listen on %s\n", cfg.Server.Address)
}
```

### Using Configuration Profiles

```go
// Load development profile
cfg, err := config.LoadProfile("development")

// Load production profile
cfg, err := config.LoadProfile("production")
```

### Loading from JSON File

```go
cfg, err := config.LoadFromFile("config.json")
```

## Configuration Structure

```json
{
  "environment": "development",
  "profile": "development",
  "server": {
    "address": ":8080",
    "path_prefix": "/api",
    "cors_origin": "*",
    "read_timeout": "10s",
    "write_timeout": "10s",
    "idle_timeout": "60s"
  },
  "storage": {
    "adapter": "memory",
    "redis": { ... },
    "sql": { ... }
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  },
  "metrics": {
    "enabled": true,
    "address": ":9090"
  }
}
```

## Environment Variables

All configuration values can be overridden using environment variables with the `GAMIFYKIT_` prefix:

```bash
export GAMIFYKIT_ENV=production
export GAMIFYKIT_SERVER_ADDR=:9090
export GAMIFYKIT_STORAGE_ADAPTER=redis
export GAMIFYKIT_LOG_LEVEL=debug
```

### Complete Environment Variable Reference

| Variable | Description | Default |
|----------|-------------|---------|
| `GAMIFYKIT_ENV` | Environment (development/testing/staging/production) | development |
| `GAMIFYKIT_PROFILE` | Configuration profile name | default |
| `GAMIFYKIT_SERVER_ADDR` | Server listen address | :8080 |
| `GAMIFYKIT_SERVER_PATH_PREFIX` | API path prefix | /api |
| `GAMIFYKIT_SERVER_CORS_ORIGIN` | CORS origin | * |
| `GAMIFYKIT_STORAGE_ADAPTER` | Storage adapter (memory/redis/sql/file) | memory |
| `GAMIFYKIT_LOG_LEVEL` | Log level (debug/info/warn/error) | info |
| `GAMIFYKIT_LOG_FORMAT` | Log format (json/text) | json |
| `GAMIFYKIT_METRICS_ENABLED` | Enable metrics collection | false |

## Configuration Profiles

### Development Profile
- Memory storage
- Debug logging
- Local server settings
- Metrics disabled

### Testing Profile
- Memory storage
- Warning level logging
- Random server port
- Metrics disabled

### Staging Profile
- File storage
- Info level logging
- Basic security features enabled

### Production Profiles
- Redis or SQL storage
- JSON structured logging
- Comprehensive security
- Metrics enabled

## Secret Management

For production deployments, sensitive configuration like database passwords should be stored in external secret stores:

```go
// Load secrets from environment (for development)
err := cfg.LoadSecretsFromEnv(ctx)

// Or use a custom secret store
store := MySecretStore{}
err := cfg.LoadSecrets(ctx, store)
```

Required secrets in production:
- `GAMIFYKIT_DATABASE_DSN` - Database connection string
- `GAMIFYKIT_REDIS_PASSWORD` - Redis password (if applicable)

## Validation

Configuration is automatically validated on load. Validation includes:

- Required fields presence
- Value ranges and formats
- Cross-field consistency
- Adapter-specific requirements

Invalid configurations will return detailed error messages indicating exactly what needs to be fixed.

## Custom Secret Stores

Implement the `SecretStore` interface for custom secret management:

```go
type SecretStore interface {
    Get(ctx context.Context, key string) (string, error)
    GetWithDefault(ctx context.Context, key, defaultValue string) string
}
```

This allows integration with:
- AWS Secrets Manager
- HashiCorp Vault
- Azure Key Vault
- Google Cloud Secret Manager
- Local encrypted files
