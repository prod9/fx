package cmdutil

import (
	"context"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/fxlog"
	"github.com/jmoiron/sqlx"
)

func NewBasicContext() (context.Context, *config.Source) {
	cfg := config.Configure()
	ctx := context.Background()
	return config.NewContext(ctx, cfg), cfg
}

func NewDataContext() (context.Context, *sqlx.DB, func()) {
	ctx, cfg := NewBasicContext()
	db := data.MustConnect(cfg)
	cleanup := func() {
		if err := db.Close(); err != nil {
			fxlog.Errorf("cmdutil: %w", err)
		}
	}
	return data.NewContext(ctx, db), db, cleanup
}

func NewMigratorContext() (context.Context, *migrator.Migrator, func()) {
	ctx, db, cleanup := NewDataContext()
	src := migrator.FromAuto(config.FromContext(ctx))
	return ctx, migrator.New(db, src), cleanup
}
