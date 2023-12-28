package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"fx.prodigy9.co/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	DatabaseURLConfig   = config.Str("DATABASE_URL")
	MigrationPathConfig = config.StrDef("DATABASE_MIGRATIONS", "./")
)

func MustConnect(cfg *config.Source) *sqlx.DB {
	if db, err := Connect(cfg); err != nil {
		log.Panicln(err)
		return nil
	} else {
		return db
	}
}

func Connect(cfg *config.Source) (*sqlx.DB, error) {
	dbURL := config.Get(cfg, DatabaseURLConfig)
	if db, err := sqlx.Open("pgx", dbURL); err != nil {
		return nil, fmt.Errorf("database: %w", err)
	} else {
		return db, nil
	}
}

func CreateDB(cfg *config.Source) error {
	rawURL := config.Get(cfg, DatabaseURLConfig)

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	dbName := parsedURL.Path
	if strings.HasPrefix(dbName, "/") {
		dbName = dbName[1:]
	}
	parsedURL.Path = "/postgres" // since our db is yet to be created

	db, err := sqlx.Open("pgx", parsedURL.String())
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	if _, err = db.Exec("CREATE DATABASE " + dbName); err != nil {
		return fmt.Errorf("database: %w", err)
	} else {
		return nil
	}
}

func NewScope(ctx context.Context, db *sqlx.DB) (Scope, error) {
	if db == nil {
		db = FromContext(ctx)
	}

	if impl, err := newScope(ctx, db); err != nil {
		return nil, err
	} else {
		return impl, nil
	}
}

func NewScopeErr(ctx context.Context, outerr *error) (Scope, context.CancelFunc, error) {
	if scope, err := NewScope(ctx, nil); err != nil {
		return nil, func() {}, err
	} else {
		return scope, func() { scope.End(outerr) }, err
	}
}

func IsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func Get(ctx context.Context, out any, sql string, args ...any) (err error) {
	return Run(ctx, func(s Scope) error {
		return s.Get(out, sql, args...)
	})
}

func Select(ctx context.Context, out any, sql string, args ...any) (err error) {
	return Run(ctx, func(s Scope) error {
		return s.Select(out, sql, args...)
	})
}

func Exec(ctx context.Context, sql string, args ...any) error {
	return Run(ctx, func(s Scope) error {
		return s.Exec(sql, args...)
	})
}

func Run(ctx context.Context, action func(s Scope) error) (err error) {
	var scope Scope
	if scope, err = NewScope(ctx, nil); err != nil {
		return
	} else {
		defer scope.End(&err)
		return action(scope)
	}
}
