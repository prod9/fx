package middlewares

import (
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxsentry"
)

// Sentry scopes each request and reports its panics when SENTRY_DSN is set; otherwise it
// passes requests through. Whether Sentry is on is decided once, when the handler is
// built.
func Sentry(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		if !fxsentry.Activate(cfg) {
			return h
		}

		return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			ctx, scope := fxsentry.NewScope(req.Context())
			scope.SetRequest(req)
			defer fxsentry.CapturePanic(ctx)

			h.ServeHTTP(resp, req.WithContext(ctx))
		})
	}
}
