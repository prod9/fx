package middlewares

import (
	"errors"
	"net/http"
	"sync/atomic"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/httpserver/render"
)

var (
	MustMigrateConfig = config.BoolDef("MUST_MIGRATE", false)

	ErrNoDB              = errors.New("database not available")
	ErrDirtyMigrations   = errors.New("database migrations are dirty")
	ErrPendingMigrations = errors.New("there are pending unapplied migrations, migrate first")
)

// CheckMigrations refuses requests while migrations are dirty or pending, when
// MUST_MIGRATE is set. Once a check finds them clean it passes every later request
// through without checking again.
func CheckMigrations(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		var clean atomic.Bool

		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			if clean.Load() {
				h.ServeHTTP(resp, req)
				return
			}

			reqCfg := config.FromContext(req.Context())
			if reqCfg == nil {
				reqCfg = config.Configure()
			}

			if !config.Get(reqCfg, MustMigrateConfig) {
				h.ServeHTTP(resp, req)
				return
			}

			db := data.FromContext(req.Context())
			if db == nil {
				render.Error(resp, req, 500, ErrNoDB)
				return
			}

			m := migrator.New(db, migrator.FromAuto(reqCfg))
			plans, dirty, err := m.Plan(req.Context(), migrator.IntentMigrate)
			switch {
			case err != nil:
				render.Error(resp, req, 500, err)
			case dirty:
				render.Error(resp, req, 500, ErrDirtyMigrations)
			case len(plans) > 0:
				render.Error(resp, req, 500, ErrPendingMigrations)
			default:
				clean.Store(true)
				h.ServeHTTP(resp, req)
			}
		})
	}
}
