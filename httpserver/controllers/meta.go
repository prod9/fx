package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"fx.prodigy9.co/buildinfo"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

// HealthzPingTimeout caps the DB ping in /healthz so a structured 503 lands well
// before kubelet's default 1s probe timeout. See
// docs/notes/2026-06-16-readiness-probe-semantics.md.
const HealthzPingTimeout = 500 * time.Millisecond

var DebugSentryConfig = config.Bool("DEBUG_SENTRY")

type meta struct {
	prefix string
}

// Meta serves the deployment-check and introspection endpoints:
//
//   - GET /                  — liveness echo, returns current server time as JSON.
//   - GET /healthz           — readiness probe. 200 when the wired-up deps are reachable;
//     503 otherwise. Currently pings the DB if one is in context.
//   - GET <prefix>/version   — the build's version as plain text.
//   - GET <prefix>/buildinfo — the build's buildinfo.Info as JSON.
//   - GET <prefix>/panic     — a test panic, mounted only when DEBUG_SENTRY is set.
func Meta(prefix string) Interface {
	return meta{prefix: prefix}
}

func (m meta) Mount(cfg *config.Source, router chi.Router) error {
	router.Get("/", m.Index)
	router.Get("/healthz", m.Healthz)
	router.Route(m.prefix, func(r chi.Router) {
		r.Get("/version", m.Version)
		r.Get("/buildinfo", m.BuildInfo)
		if config.Get(cfg, DebugSentryConfig) {
			r.Get("/panic", m.Panic)
		}
	})
	return nil
}

func (m meta) Index(resp http.ResponseWriter, r *http.Request) {
	render.JSON(resp, r, struct {
		Time time.Time `json:"time"`
	}{time.Now()})
}

func (m meta) Healthz(resp http.ResponseWriter, r *http.Request) {
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

func (m meta) Version(resp http.ResponseWriter, r *http.Request) {
	render.Text(resp, r, buildinfo.FromContext(r.Context()).Version)
}

func (m meta) BuildInfo(resp http.ResponseWriter, r *http.Request) {
	render.JSON(resp, r, buildinfo.FromContext(r.Context()))
}

func (m meta) Panic(resp http.ResponseWriter, r *http.Request) {
	panic("this is a test panic")
}
