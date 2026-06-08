package posts

import (
	"embed"

	"fx.prodigy9.co/app"
)

//go:embed *.up.sql *.down.sql
var migrationsFS embed.FS

var App = app.Build().
	Name("posts").
	Description("Posts fragment — owns posts table, FK to users").
	EmbedMigrations(migrationsFS)
