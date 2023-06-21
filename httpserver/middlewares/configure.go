package middlewares

import (
	"fx.prodigy9.co/config"
	"net/http"
)

func Configure(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(resp, config.NewRequest(r, cfg))
		})
	}
}
