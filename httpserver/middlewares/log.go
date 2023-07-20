package middlewares

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"fx.prodigy9.co/config"
	"github.com/felixge/httpsnoop"
)

var EnableRequestLog = config.BoolDef("LOG_REQUESTS", true)

func LogRequests(cfg *config.Source) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
			if !config.Get(cfg, EnableRequestLog) {
				h.ServeHTTP(resp, r)
				return
			}

			var metrics *httpsnoop.Metrics
			log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.RequestURI)
			defer func() {
				if metrics == nil {
					return
				}
				body := r.Context().Value("reqBody")
				if body == nil {
					body = ""
				}
				log.Printf("%s %s %s %s - HTTP %d %s\n",
					r.RemoteAddr, r.Method, r.RequestURI, body,
					metrics.Code, metrics.Duration)
			}()

			m := httpsnoop.CaptureMetrics(h, resp, r)
			metrics = &m
		})
	}
}

func CopyRequestBody(cfg *config.Source) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			bodyBytes, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			ctx := context.WithValue(req.Context(), "reqBody", string(bodyBytes))
			next.ServeHTTP(rw, req.WithContext(ctx))
		})
	}
}
