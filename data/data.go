package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/dbname"
	"fx.prodigy9.co/fxlog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	DatabaseURLConfig     = config.Str("DATABASE_URL")
	DatabaseMaxIdleConfig = config.IntDef("DATABASE_MAX_IDLE", runtime.NumCPU())
	DatabaseMaxOpenConfig = config.IntDef("DATABASE_MAX_OPEN", -1)
)

func MustConnect(cfg *config.Source) *sqlx.DB {
	if db, err := Connect(cfg); err != nil {
		fxlog.Fatalf("data: connection failed: %w", err)
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
		configureDB(cfg, db)
		return db, nil
	}
}
func configureDB(cfg *config.Source, db *sqlx.DB) {
	maxIdle, maxOpen :=
		config.Get(cfg, DatabaseMaxIdleConfig),
		config.Get(cfg, DatabaseMaxOpenConfig)
	if maxIdle > 0 {
		db.SetMaxIdleConns(maxIdle)
	}
	if maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	}
}

func CreateDB(cfg *config.Source) error {
	conn, dbName, err := getAdminConnection(cfg)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	if _, err := conn.Exec("CREATE DATABASE " + dbName); err != nil {
		return fmt.Errorf("database: %w", err)
	} else {
		return nil
	}
}

func DropDB(cfg *config.Source) error {
	conn, dbName, err := getAdminConnection(cfg)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	if _, err := conn.Exec("DROP DATABASE " + dbName); err != nil {
		return fmt.Errorf("database: %w", err)
	} else {
		return nil
	}
}

func getAdminConnection(cfg *config.Source) (*sqlx.DB, string, error) {
	rawURL := config.Get(cfg, DatabaseURLConfig)

	name, err := dbname.From(rawURL)
	if err != nil {
		return nil, "", err
	}

	defDBURL, err := dbname.SetDefaultDB(rawURL)
	if err != nil {
		return nil, "", err
	}

	if db, err := sqlx.Open("pgx", defDBURL); err != nil {
		return nil, "", err
	} else {
		return db, name, nil
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
