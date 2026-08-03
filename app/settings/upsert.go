package settings

import (
	"net/http"

	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type UpsertBody struct {
	Value string `json:"value"`
}

func (c Ctr) Upsert(resp http.ResponseWriter, req *http.Request) {
	slug := chi.URLParam(req, "slug")
	st := &UpsertBody{}
	if err := controllers.ReadJSON(req, st); err != nil {
		render.Error(resp, req, 400, err)
		return
	}

	if settings, err := Set(req.Context(), slug, st.Value); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, settings)
	}
}
