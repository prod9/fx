# Database Migrations

**Status:** accepted

A built-in migration engine is provided in `data/migrator`. To use this, add data
commands to the application:

```go
app := app.Build().
  Name("my todo app").
  // or just .AddDefaults() which already includes data commands
  Commands(cmd.BuildDataCommand()).
```

`cmd.BuildDataCommand(srcs ...migrator.Source)` builds the `data` command group.
Explicit sources — for example, migrations collected from an app tree — are threaded
into every migration-reading subcommand (`migrate`, `rollback`, `list-migrations`,
`collect-migrations`, `resync-migrations`); called with none, the commands use the
auto-detected sources alone.

The following commands become available:

* `go run . data migrate` — Runs all migrations.
* `go run . data new-migration (name) [subdir]` — Creates new up+down migration files.

Migrations are written as normal SQL files. Usually they contain `CREATE TABLE` for the
up migration and `DROP TABLE` for the down migration.

During production deployment, migrations can be collected and embedded into the
application itself using the `go:embed` directive for easy distribution:

```go
//go:embed */*.sql
var sqlMigrations embed.FS

func main() {
  err := app.Build().
    AddDefaults().
    EmbedMigrations(sqlMigrations). // <-- Add this line

    // ...

    Start()

  if err != nil {
    log.Fatalln(err)
  }
}
```

Migration discovery (`migrator.LoadAuto`) resolves in tiers — first non-empty tier
wins:

1. The path configured via `DATABASE_MIGRATIONS`, if set (empty is an error — it is
   likely a misconfiguration).
2. The current working directory, recursively.
3. The embedded tier: sources registered via `migrator.Embed` plus any sources passed
   explicitly (e.g. through `cmd.BuildDataCommand`), union-merged.

The union-merge itself is `migrator.Collect(srcs ...migrator.Source)` — it loads every
source, skips those with no migrations, sorts the union by name, and returns
`ErrNoMigrations` when the union is empty. Use it directly when composing migrations
outside the auto path.

Use the `data list-migrations` command to check what is detected:

```sh
$ go run ./api data list-migrations

api/auth/202312281812_create_users_and_sessions.up.sql
api/listing/202504011719_create_listing.up.sql
api/files/202504041033_create_files.up.sql
```

App fragments that ship their own embedded migrations (`files.App`, `settings.App`,
etc.) are aggregated automatically when mounted — you do not need to re-embed them at
the root.

Other commands include:

* `go run . data collect-migrations (outdir)` — Collect migration files into a single
  directory.
* `go run . data create-db` — Creates database specified in the config.
* `go run . data list-migrations` — List all detected migration files.
* `go run . data migrate` — Runs all detected migration scripts.
* `go run . data new-migration (name) [subdir]` — Creates new up+down migration files.
* `go run . data psql` — Starts a psql shell connecting to the configured database.
* `go run . data recover-migrations [output-dir]` — Export migration cache from
  database to files.
* `go run . data resync-migrations` — Update database migration cache to match program
  files.
* `go run . data rollback` — Revert one previously run migration.

## Scripting and CI

Data commands use the `cmd/prompts` package for interactive input. To run commands
non-interactively (e.g. in CI/CD pipelines or scripts), set `CI=1`. In CI mode, all
required inputs must be provided as positional arguments — if any are missing, the
command will exit with an error instead of prompting.

Additionally, set `ALWAYS_YES=1` to automatically confirm all yes/no prompts.
