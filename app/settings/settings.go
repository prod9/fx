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

var App = app.Build().
	EmbedMigrations(migrations).
	Controllers(Ctr{})

type Settings struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func List(ctx context.Context) ([]*Settings, error) {
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

// TODO: Cache
func Get(ctx context.Context, key string) (*Settings, error) {
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

func Set(ctx context.Context, key string, value string) (*Settings, error) {
	const sql = `
	UPDATE settings
	SET value = $2,
		updated_at = $3
	WHERE key = $1
	RETURNING *
	`

	settings := &Settings{}
	if err := data.Get(ctx, settings, sql, key, value, time.Now()); err != nil {
		return nil, err
	} else {
		return settings, nil
	}
}

func Delete(ctx context.Context, key string) (*Settings, error) {
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
