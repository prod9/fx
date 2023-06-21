package data

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type dbContextKey struct{}

func NewContext(ctx context.Context, db *sqlx.DB) context.Context {
	return context.WithValue(ctx, dbContextKey{}, db)
}

func FromContext(ctx context.Context) *sqlx.DB {
	return ctx.Value(dbContextKey{}).(*sqlx.DB)
}
