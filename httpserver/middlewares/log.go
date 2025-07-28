package middlewares

import (
	"net/http"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxlog"
	"github.com/felixge/httpsnoop"
)

// NOTE: go-chi also provides https://github.com/go-chi/httplog, might worth investigating
func LogRequests(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
			var metrics *httpsnoop.Metrics
			fxlog.Log("request",
				fxlog.String("remote", r.RemoteAddr),
				fxlog.String("method", r.Method),
				fxlog.String("uri", r.RequestURI),
			)

			defer func() {
				if metrics == nil {
					return
				}

				fxlog.Log("response",
					fxlog.String("remote", r.RemoteAddr),
					fxlog.String("method", r.Method),
					fxlog.String("uri", r.RequestURI),

					fxlog.Int("code", metrics.Code),
					fxlog.Duration("duration", metrics.Duration),
					fxlog.Int64("bytes", metrics.Written),
				)
			}()

			m := httpsnoop.CaptureMetrics(h, resp, r)
			metrics = &m
		})
	}
}
