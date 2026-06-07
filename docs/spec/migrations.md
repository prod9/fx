# Database Migrations

**Status:** accepted

A built-in migration engine is provided in `data/migrator`. To use this, add data
commands to the application:

```go
app := app.Build().
  Name("my todo app").
  // or just .AddDefaults() which already includes data commands
  Commands(data.Cmd).
```

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

The migrator will automatically look for embedded sources. Otherwise it will look at
the current folder, and its parents, for the files. Use the `data list-migrations`
command to check:

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
