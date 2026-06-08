package comments

import (
	"embed"

	"fx.prodigy9.co/app"
)

//go:embed *.up.sql *.down.sql
var migrationsFS embed.FS

var App = app.Build().
	Name("comments").
	Description("Comments fragment — owns comments table, FK to posts").
	EmbedMigrations(migrationsFS)
