package settings

import (
	"context"
	"embed"
	"time"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/data"
)

//go:embed *.sql
var migrations embed.FS

// CreateSettingsTableSQL is idempotent (IF NOT EXISTS), so the entry primitives call it
// lazily via ensureSettingsTable — a fresh database works from any context with no prior
// migration. The embedded migration below carries the same schema for deploys that run
// `data migrate`; the two are belt-and-suspenders, never in conflict.
const CreateSettingsTableSQL = `
CREATE TABLE IF NOT EXISTS settings (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL,

	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

// App is the settings fragment: the settings table (self-initialized on first access, or
// via the embedded migration on deploy). Its REST controller is deliberately not a
// fragment auto-mount — call settings.Mount inside a route group you own and guard, so
// settings data is never exposed by the mere act of mounting App.
var App = app.Build().
	EmbedMigrations(migrations)

type Settings struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func List(ctx context.Context) ([]*Settings, error) {
	if err := ensureSettingsTable(ctx); err != nil {
		return nil, err
	}

	const sql = `
	SELECT *
	FROM settings
	ORDER BY created_at ASC
	`

	var settings []*Settings
	if err := data.Select(ctx, &settings, sql); err != nil {
		return nil, err
	} else {
		return settings, nil
	}
}

// Get returns the value stored for key, or fallback when no row exists. Absence is not an
// error; only a real query failure is. Callers that must distinguish absence use lookup.
//
// TODO: Cache
func Get(ctx context.Context, key, fallback string) (string, error) {
	setting, err := lookup(ctx, key)
	switch {
	case data.IsNoRows(err):
		return fallback, nil
	case err != nil:
		return "", err
	default:
		return setting.Value, nil
	}
}

func Delete(ctx context.Context, key string) (*Settings, error) {
	if err := ensureSettingsTable(ctx); err != nil {
		return nil, err
	}

	const sql = `
	DELETE FROM settings
	WHERE key = $1
	RETURNING *
	`

	settings := &Settings{}
	if err := data.Get(ctx, settings, sql, key); err != nil {
		return nil, err
	} else {
		return settings, nil
	}
}

// lookup returns the row for key, or IsNoRows if absent. It is the found/not-found
// primitive behind Get and config.Provider.Get.
func lookup(ctx context.Context, key string) (*Settings, error) {
	if err := ensureSettingsTable(ctx); err != nil {
		return nil, err
	}

	const sql = `
	SELECT *
	FROM settings
	WHERE key = $1
	ORDER BY created_at ASC
	LIMIT 1;
	`

	settings := &Settings{}
	if err := data.Get(ctx, settings, sql, key); err != nil {
		return nil, err
	} else {
		return settings, nil
	}
}

func ensureSettingsTable(ctx context.Context) error {
	return data.Exec(ctx, CreateSettingsTableSQL)
}
