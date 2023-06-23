package coda

import (
	"context"

	"fx.prodigy9.co/config"
)

type contextKey struct{}

func FromContext(ctx context.Context) *Client {
	return ctx.Value(contextKey{}).(*Client)
}

func FromContextOrNew(ctx context.Context, cfg *config.Source) *Client {
	if client, ok := ctx.Value(contextKey{}).(*Client); ok {
		return client
	} else {
		if cfg == nil {
			cfg = config.FromContext(ctx)
		}
		return NewClient(cfg)
	}
}

func NewContext(ctx context.Context, client *Client) context.Context {
	if client == nil {
		cfg := config.FromContext(ctx)
		client = NewClient(cfg)
	}

	return context.WithValue(ctx, contextKey{}, client)
}
