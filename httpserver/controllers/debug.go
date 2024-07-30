package controllers

import (
	"net/http"

	"fx.prodigy9.co/config"
	"github.com/go-chi/chi/v5"
)

// Debug is a utility controller providing some utility routes for testing and checking
// the system.
type Debug struct{}

var _ Interface = Debug{}

func (h Debug) Mount(cfg *config.Source, router chi.Router) error {
	router.Get("/__panic", h.Panic)
	return nil
}

func (h Debug) Panic(resp http.ResponseWriter, r *http.Request) {
	panic("this is a test panic")
}
