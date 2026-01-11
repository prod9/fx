package cmdutil

import (
	"context"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"github.com/jmoiron/sqlx"
)

func NewBasicContext() (context.Context, *config.Source) {
	cfg := config.Configure()
	ctx := context.Background()
	return config.NewContext(ctx, cfg), cfg
}

func NewDataContext() (context.Context, *sqlx.DB) {
	ctx, cfg := NewBasicContext()
	db := data.MustConnect(cfg)
	return data.NewContext(ctx, db), db
}

func NewMigratorContext() (context.Context, *migrator.Migrator) {
	ctx, db := NewDataContext()
	src := migrator.FromAuto(config.FromContext(ctx))
	return ctx, migrator.New(db, src)
}
