// Package audit records an append-only trail of actions to a Postgres audit_events
// table. Mount App so the migration ships with the service — importing the package for
// Log/Record alone does not create the table. Actions are entity.verb_past strings:
// services define their own constants (fx does not enumerate them), and target_type is
// derived from the action prefix.
package audit

import (
	"embed"

	"fx.prodigy9.co/app"
)

//go:embed *.sql
var migrations embed.FS

// App is the mountable fragment carrying the audit_events migration.
var App = app.Build().EmbedMigrations(migrations)
