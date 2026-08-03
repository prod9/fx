package settings

import (
	"context"
	"net/http"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/render"
	"github.com/go-chi/chi/v5"
)

type Upsert struct {
	Key   string `json:"-"`
	Value string `json:"value"`
}

var _ controllers.Action = (*Upsert)(nil)

func (a *Upsert) Execute(ctx context.Context, out any) error {
	if err := ensureSettingsTable(ctx); err != nil {
		return err
	}
	return data.Get(ctx, out, `
		INSERT INTO settings (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE
		SET value = $2,
			updated_at = CURRENT_TIMESTAMP
		RETURNING *`,
		a.Key, a.Value)
}

func (c ctr) Upsert(resp http.ResponseWriter, req *http.Request) {
	action, setting := &Upsert{Key: chi.URLParam(req, "slug")}, &Settings{}
	if err := controllers.ExecuteAction(resp, req, action, setting); err != nil {
		render.Error(resp, req, 500, err)
	} else {
		render.JSON(resp, req, setting)
	}
}
