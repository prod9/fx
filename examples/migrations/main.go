// Demonstrates fragment migration aggregation.
//
// Three child fragments (users, posts, comments) each register their own embedded
// migrations via app.Builder.EmbedMigrations. The root app does not embed any
// migrations itself. When `data migrate` runs, app.Start walks the child tree and
// the migrator picks up all three sets, sorted by filename (timestamp-prefixed).
//
// Run:
//
//	export DATABASE_URL=postgres://localhost/fx_migrations_example
//	go run . data create-db
//	go run . data migrate
//	go run . data rollback   # x3 to walk back down
package main

import (
	"log"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/examples/migrations/comments"
	"fx.prodigy9.co/examples/migrations/posts"
	"fx.prodigy9.co/examples/migrations/users"
)

func main() {
	err := app.Build().
		Name("migrations-example").
		Description("Demonstrates fragment migration aggregation").
		AddDefaults().
		Mount(users.App).
		Mount(posts.App).
		Mount(comments.App).
		Start()

	if err != nil {
		log.Fatalln(err)
	}
}
