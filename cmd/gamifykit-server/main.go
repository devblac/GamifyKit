package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	mem "gamifykit/adapters/memory"
	redisAdapter "gamifykit/adapters/redis"
	sqlxAdapter "gamifykit/adapters/sqlx"
	"gamifykit/api/httpapi"
	"gamifykit/config"
	"gamifykit/engine"
	"gamifykit/gamify"
	"gamifykit/realtime"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logging based on configuration
	setupLogging(cfg)

	// Load secrets if in production
	ctx := context.Background()
	if cfg.Environment == config.EnvProduction {
		if err := cfg.LoadSecretsFromEnv(ctx); err != nil {
			slog.Error("Failed to load secrets", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("starting gamifykit server",
		"environment", cfg.Environment,
		"profile", cfg.Profile,
		"address", cfg.Server.Address,
		"storage_adapter", cfg.Storage.Adapter)

	// Setup storage adapter
	storage, err := setupStorage(ctx, cfg)
	if err != nil {
		slog.Error("Failed to setup storage", "error", err)
		os.Exit(1)
	}

	// Build service
	hub := realtime.NewHub()
	svc := gamify.New(
		gamify.WithRealtime(hub),
		gamify.WithStorage(storage),
		gamify.WithDispatchMode(engine.DispatchAsync),
	)

	// Setup HTTP API
	handler := httpapi.NewMux(svc, hub, httpapi.Options{
		PathPrefix:      cfg.Server.PathPrefix,
		AllowCORSOrigin: cfg.Server.CORSOrigin,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("server listening", "address", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server", "timeout", cfg.Server.ShutdownTimeout)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during server shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

// setupLogging configures the logger based on configuration
func setupLogging(cfg *config.Config) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Logging.Level),
	}

	switch cfg.Logging.Format {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	// Add attributes if specified
	if len(cfg.Logging.Attributes) > 0 {
		handler = handler.WithAttrs(convertAttributes(cfg.Logging.Attributes))
	}

	slog.SetDefault(slog.New(handler))
}

// parseLogLevel converts string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// convertAttributes converts map[string]string to []slog.Attr
func convertAttributes(attrs map[string]string) []slog.Attr {
	var result []slog.Attr
	for k, v := range attrs {
		result = append(result, slog.String(k, v))
	}
	return result
}

// setupStorage creates the appropriate storage adapter based on configuration
func setupStorage(ctx context.Context, cfg *config.Config) (engine.Storage, error) {
	switch cfg.Storage.Adapter {
	case "memory":
		return mem.New(), nil

	case "redis":
		return redisAdapter.New(cfg.Storage.Redis)

	case "sql":
		return sqlxAdapter.New(cfg.Storage.SQL)

	case "file":
		return mem.New(), fmt.Errorf("file storage not yet implemented, using memory fallback")

	default:
		return mem.New(), fmt.Errorf("unknown storage adapter: %s", cfg.Storage.Adapter)
	}
}
