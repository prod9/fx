package drops

import (
	"context"
	"embed"
	"time"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/data"
)

//go:embed *.sql
var migrations embed.FS

// App is the drops fragment: a public "send a file to a friend" share. Each drop
// owns a single file, addressed by an unguessable token.
var App = app.Build().
	Description("Send-a-file-to-a-friend drops").
	EmbedMigrations(migrations).
	Controllers(Ctr{})

type Drop struct {
	ID        int64     `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func GetDropByToken(ctx context.Context, token string) (*Drop, error) {
	drop := &Drop{}
	sql := `SELECT * FROM drops WHERE token = $1 LIMIT 1`
	if err := data.Get(ctx, drop, sql, token); err != nil {
		return nil, err
	}
	return drop, nil
}
