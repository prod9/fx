package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

// HealthzPingTimeout caps the DB ping in /healthz so a structured 503 lands well
// before kubelet's default 1s probe timeout. See
// docs/notes/2026-06-16-readiness-probe-semantics.md.
const HealthzPingTimeout = 500 * time.Millisecond

// Home serves the basic deployment-check endpoints.
//
//   - GET /        — liveness echo, returns current server time as JSON.
//   - GET /healthz — readiness probe. 200 when the wired-up deps are reachable;
//     503 otherwise. Currently pings the DB if one is in context.
type Home struct{}

var _ Interface = Home{}

func (h Home) Mount(cfg *config.Source, router chi.Router) error {
	router.Get("/", h.Index)
	router.Get("/healthz", h.Healthz)
	return nil
}

func (h Home) Index(resp http.ResponseWriter, r *http.Request) {
	render.JSON(resp, r, struct {
		Time time.Time `json:"time"`
	}{time.Now()})
}

func (h Home) Healthz(resp http.ResponseWriter, r *http.Request) {
	db, ok := data.LookupFromContext(r.Context())
	if !ok {
		render.JSON(resp, r, map[string]string{"status": "ok"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), HealthzPingTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		render.Error(resp, r, http.StatusServiceUnavailable,
			fmt.Errorf("database unreachable: %w", err))
		return
	}

	render.JSON(resp, r, map[string]string{"status": "ok", "db": "ok"})
}
