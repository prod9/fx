package middlewares

import (
	"fx.prodigy9.co/config"
	"github.com/rs/cors"
	"net/http"
	"strings"
)

var CORSOriginConfig = config.Str("API_CORS_ORIGINS")

func CORS(cfg *config.Source, options cors.Options) func(http.Handler) http.Handler {
	origin := strings.TrimSpace(config.Get(cfg, CORSOriginConfig))
	if origin != "" {
		options.AllowedOrigins = strings.Split(origin, ",")
	}

	return cors.New(options).Handler
}

func CORSAllowAll(cfg *config.Source) func(http.Handler) http.Handler {
	var handler *cors.Cors

	origin := strings.TrimSpace(config.Get(cfg, CORSOriginConfig))
	if origin == "" {
		handler = cors.AllowAll()
	} else {
		handler = cors.New(cors.Options{
			AllowedOrigins:   strings.Split(origin, ","),
			AllowCredentials: true, // support fetch() with {credentials: 'include'}
			AllowedHeaders:   []string{"*"},
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
		})
	}

	return handler.Handler
}
