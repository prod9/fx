package settings

import (
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/httperrors"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type ctr struct{}

// Controller serves the settings endpoints under /settings. It is not a fragment
// auto-mount: mount it inside a route group you define and guard (auth / RBAC / IP
// allowlist — your call), so settings data is never exposed by mounting settings.App
// alone.
func Controller() controllers.Interface {
	return ctr{}
}

func (c ctr) Mount(cfg *config.Source, router chi.Router) error {
	router.Route("/settings", func(r chi.Router) {
		r.Get("/", c.Index)
		r.Post("/{slug}", c.Upsert)
		r.Delete("/{slug}", c.Delete)
	})
	return nil
}

func (c ctr) Index(resp http.ResponseWriter, req *http.Request) {
	if settings, err := List(req.Context()); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, settings)
	}
}

func (c ctr) Delete(resp http.ResponseWriter, req *http.Request) {
	slug := chi.URLParam(req, "slug")
	if settings, err := Delete(req.Context(), slug); data.IsNoRows(err) {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
	} else if err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, settings)
	}
}
