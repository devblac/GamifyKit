package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	wsadapter "gamifykit/adapters/websocket"
	"gamifykit/core"
	"gamifykit/engine"
	"gamifykit/realtime"
)

// Options configures the HTTP API surface.
type Options struct {
	// PathPrefix, if set, is prepended to all routes (e.g., "/api").
	PathPrefix string
	// AllowCORSOrigin, if non-empty, enables basic CORS with the given origin (use "*" for any).
	AllowCORSOrigin string
}

// NewMux builds an http.Handler exposing a minimal Gamify REST API and WebSocket stream.
// Routes:
//   - POST {prefix}/users/{id}/points?metric=xp&delta=50
//   - POST {prefix}/users/{id}/badges/{badge}
//   - GET  {prefix}/users/{id}
//   - GET  {prefix}/healthz
//   - WS   {prefix}/ws
func NewMux(svc *engine.GamifyService, hub *realtime.Hub, opts Options) http.Handler {
	mux := http.NewServeMux()

	// health
	mux.HandleFunc(withPrefix(opts.PathPrefix, "/healthz"), func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true})
	})

	// WebSocket events
	if hub != nil {
		mux.Handle(withPrefix(opts.PathPrefix, "/ws"), wsadapter.Handler(hub))
	}

	// Users API
	mux.HandleFunc(withPrefix(opts.PathPrefix, "/users/"), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		parts := split(r.URL.Path, '/')
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		user := core.UserID(parts[1])
		switch r.Method {
		case http.MethodPost:
			if len(parts) >= 3 && parts[2] == "points" {
				metric := core.Metric(r.URL.Query().Get("metric"))
				if metric == "" {
					metric = core.MetricXP
				}
				delta, _ := strconv.ParseInt(r.URL.Query().Get("delta"), 10, 64)
				total, err := svc.AddPoints(r.Context(), user, metric, delta)
				writeJSON(w, map[string]any{"total": total, "err": errString(err)})
				return
			}
			if len(parts) >= 4 && parts[2] == "badges" {
				badge := core.Badge(parts[3])
				err := svc.AwardBadge(r.Context(), user, badge)
				writeJSON(w, map[string]any{"ok": err == nil, "err": errString(err)})
				return
			}
		case http.MethodGet:
			st, err := svc.GetState(r.Context(), user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, st)
			return
		}
		http.NotFound(w, r)
	})

	var handler http.Handler = mux
	if opts.AllowCORSOrigin != "" {
		handler = withCORS(handler, opts.AllowCORSOrigin)
	}
	return handler
}

// Helpers

func withPrefix(prefix, path string) string {
	if prefix == "" || prefix == "/" {
		return path
	}
	if prefix[len(prefix)-1] == '/' {
		return prefix[:len(prefix)-1] + path
	}
	return prefix + path
}

func split(p string, sep rune) []string {
	var parts []string
	cur := make([]rune, 0, len(p))
	// trim leading '/'
	for len(p) > 0 && p[0] == '/' {
		p = p[1:]
	}
	for _, r := range p {
		if r == sep {
			if len(cur) > 0 {
				parts = append(parts, string(cur))
				cur = cur[:0]
			}
			continue
		}
		cur = append(cur, r)
	}
	if len(cur) > 0 {
		parts = append(parts, string(cur))
	}
	return parts
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func errString(err error) any {
	if err == nil {
		return nil
	}
	return err.Error()
}

// withCORS wraps a handler with a minimal CORS policy.
func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
