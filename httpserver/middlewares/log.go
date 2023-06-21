package middlewares

import (
	"log"
	"net/http"

	"fx.prodigy9.co/config"
	"github.com/felixge/httpsnoop"
)

func LogRequests(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
			var metrics *httpsnoop.Metrics
			log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.RequestURI)
			defer func() {
				if metrics == nil {
					return
				}

				log.Printf("%s %s %s - HTTP %d %s\n",
					r.RemoteAddr, r.Method, r.RequestURI,
					metrics.Code, metrics.Duration)
			}()

			m := httpsnoop.CaptureMetrics(h, resp, r)
			metrics = &m
		})
	}
}
