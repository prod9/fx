package users

import (
	"embed"

	"fx.prodigy9.co/app"
)

//go:embed *.up.sql *.down.sql
var migrationsFS embed.FS

var App = app.Build().
	Name("users").
	Description("Users fragment — owns users table").
	EmbedMigrations(migrationsFS)
