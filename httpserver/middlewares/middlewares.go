package middlewares

import (
	"net/http"

	"fx.prodigy9.co/config"
)

type Interface func(*config.Source) func(http.Handler) http.Handler

func DefaultForAPI() []Interface {
	return []Interface{
		Configure,
		LogRequests,
		CORSAllowAll,
		AddDataContext,
		// MountControllers middleware are called when building serve_cmd
	}
}
