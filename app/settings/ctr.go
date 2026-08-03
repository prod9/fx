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

type Ctr struct{}

var _ controllers.Interface = Ctr{}

func (c Ctr) Mount(cfg *config.Source, router chi.Router) error {
	router.Route("/settings", func(r chi.Router) {
		r.Get("/", c.Index)
		r.Post("/{slug}", c.Upsert)
		r.Delete("/{slug}", c.Delete)
	})
	return nil
}

func (c Ctr) Index(resp http.ResponseWriter, req *http.Request) {
	if settings, err := List(req.Context()); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, settings)
	}
}

func (c Ctr) Delete(resp http.ResponseWriter, req *http.Request) {
	slug := chi.URLParam(req, "slug")
	if settings, err := Delete(req.Context(), slug); data.IsNoRows(err) {
		render.Error(resp, req, 404, httperrors.ErrNotFound)
	} else if err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, settings)
	}
}
