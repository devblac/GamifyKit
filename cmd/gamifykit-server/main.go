package main

import (
	"context"
	"flag"
	"log"
	"net/http"
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

	log.Printf("gamifykit server listening on %s%s", *addr, *prefix)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	_ = srv.Shutdown(context.Background())
}
