package cmd

import (
	"context"
	"log"

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
	db, err := data.Connect(config.FromContext(ctx))
	if err != nil {
		log.Fatalln(err)
	}
	return data.NewContext(ctx, db)
}
