package config

import (
	"context"
	"net/http"
)

// for use as unique context keys
type contextKey struct{}

func NewContext(ctx context.Context, cfg *Source) context.Context {
	if cfg == nil {
		return ctx
	} else {
		return context.WithValue(ctx, contextKey{}, cfg)
	}
}

func FromContext(ctx context.Context) *Source {
	if src, ok := ctx.Value(contextKey{}).(*Source); ok {
		return src
	} else {
		return nil
	}
}

func NewRequest(r *http.Request, cfg *Source) *http.Request {
	return r.WithContext(NewContext(r.Context(), cfg))
}

func FromRequest(r *http.Request) *Source {
	return FromContext(r.Context())
}
