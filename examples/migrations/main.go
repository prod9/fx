// Demonstrates fragment migration aggregation.
//
// Four child fragments (files, users, posts, comments) each register their own
// embedded migrations via app.Builder.EmbedMigrations. The root app does not embed
// any migrations itself. When `data migrate` runs, app.Start walks the child tree
// and the migrator picks up all four sets, sorted by filename (timestamp-prefixed).
//
// `files.App` is FX's own built-in fragment — including it proves that built-in
// fragments aggregate through the same Mount path as user-defined ones.
//
// Run:
//
//	export DATABASE_URL=postgres://localhost/fx_migrations_example
//	export ALWAYS_YES=1
//	go run . data create-db
//	go run . data migrate
//	go run . data rollback   # x4 to walk back down
package main

import (
	"log"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/app/files"
	"fx.prodigy9.co/examples/migrations/comments"
	"fx.prodigy9.co/examples/migrations/posts"
	"fx.prodigy9.co/examples/migrations/users"
)

func main() {
	err := app.Build().
		Name("migrations-example").
		Description("Demonstrates fragment migration aggregation").
		AddDefaults().
		Mount(files.App).
		Mount(users.App).
		Mount(posts.App).
		Mount(comments.App).
		Start()

	if err != nil {
		log.Fatalln(err)
	}
}
