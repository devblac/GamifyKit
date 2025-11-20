package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	mem "gamifykit/adapters/memory"
	"gamifykit/api/httpapi"
	"gamifykit/engine"
	"gamifykit/gamify"
	"gamifykit/realtime"
)

func main() {
	var (
		addr   = flag.String("addr", ":8080", "listen address")
		prefix = flag.String("prefix", "/api", "HTTP API path prefix")
		cors   = flag.String("cors", "*", "Access-Control-Allow-Origin value (empty to disable)")
	)
	flag.Parse()

	// Configure JSON logging for production use
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(logHandler))

	// Build service with sensible defaults.
	hub := realtime.NewHub()
	svc := gamify.New(
		gamify.WithRealtime(hub),
		gamify.WithStorage(mem.New()),
		gamify.WithDispatchMode(engine.DispatchAsync),
	)

	// HTTP API
	handler := httpapi.NewMux(svc, hub, httpapi.Options{PathPrefix: *prefix, AllowCORSOrigin: *cors})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("starting gamifykit server",
		"address", *addr,
		"api_prefix", *prefix,
		"cors_origin", *cors)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown - this will only be reached if the server is stopped externally
	slog.Info("server shutting down")
	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("error during server shutdown", "error", err)
		os.Exit(1)
	}
}
