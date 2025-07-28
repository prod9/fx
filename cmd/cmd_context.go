package cmd

import (
	"context"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
)

func NewBasicContext() context.Context {
	cfg := config.Configure()
	ctx := context.Background()
	return config.NewContext(ctx, cfg)
}

func NewDataContext() context.Context {
	ctx := NewBasicContext()
	db := data.MustConnect(config.FromContext(ctx))
	return data.NewContext(ctx, db)
}
