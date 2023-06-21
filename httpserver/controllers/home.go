package controllers

import (
	"net/http"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

// Home is a bog-standard controller that just returns the current server's time, useful
// for getting some basic http response to test the deployment when setting up the
// application and for basic health checks.
//
// TODO: A built-in /healthz for a more involved health check (e.g. ping database)
type Home struct{}

var _ Interface = Home{}

func (h Home) Mount(cfg *config.Source, router chi.Router) error {
	router.Get("/", h.Index)
	return nil
}

func (h Home) Index(resp http.ResponseWriter, r *http.Request) {
	render.JSON(resp, r, struct {
		Time time.Time `json:"time"`
	}{time.Now()})
}
