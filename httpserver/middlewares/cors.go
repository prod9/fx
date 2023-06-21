package middlewares

import (
	"fx.prodigy9.co/config"
	"github.com/rs/cors"
	"net/http"
)

func CORSAllowAll(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return cors.AllowAll().Handler(h)
	}
}
