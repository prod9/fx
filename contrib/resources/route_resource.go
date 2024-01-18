package resources

import (
	"context"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver/render"
	"net/http"
)

// RouteResourceProvider adds a resource provider to the map of providers
func RouteResourceProviderMiddleware(key string, provider ResourceProvider) func(cfg *config.Source) func(http.Handler) http.Handler {
	return func(cfg *config.Source) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				providers, _ := req.Context().Value("resourceProviders").(map[string]ResourceProvider)
				if providers == nil {
					providers = make(map[string]ResourceProvider, 1)
				}
				providers[key] = provider
				ctx := context.WithValue(req.Context(), "resourceProviders", providers)
				next.ServeHTTP(rw, req.WithContext(ctx))
			})
		}
	}
}

func RouteResourceMapper() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			resourceProviderMap, _ := req.Context().Value("resourceProviders").(map[string]ResourceProvider)
			if resourceProviderMap != nil {
				err := MapResourcesFromRoute(req.Context(), resourceProviderMap)
				if err != nil {
					render.Error(rw, req, 404, err)
					return
				}
			}
			next.ServeHTTP(rw, req)
		})
	}
}
