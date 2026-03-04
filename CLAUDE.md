# FX — PRODIGY9 Go API Framework

Module: `fx.prodigy9.co` | Go 1.24 | Maintainer: Chakrit Wichian

FX is a minimalistic, modular Go API framework. It bundles well-integrated tools for
building APIs while letting engineers swap pieces in/out. It is designed to be used via
`git subtree` in other projects.

## ACE / Coding School

This project's AI coding environment is managed by [ACE](https://github.com/prod9/ace).
Run `ace` to start a coding session. Run `ace setup` if not yet configured.

Skills and conventions are provided by the **PRODIGY9 Coding School** school and are symlinked into
`.claude/skills/`. Skill edits go through symlinks into the school cache — propose
changes back to the school repo when ready. Run `ace config` or `ace paths` to debug
configuration issues.

## Package Map

| Package                  | Purpose                                                                    |
|--------------------------|----------------------------------------------------------------------------|
| `app`                    | Application builder, fragment composition, lifecycle                       |
| `app/files`              | S3-backed file management (presigned URLs, PostgreSQL metadata)            |
| `config`                 | Type-safe env var config (`Var[T]`, `config.Get`, `.env`/`.env.local`)     |
| `httpserver`             | HTTP server startup                                                        |
| `httpserver/controllers` | Controller interface (`Mount(cfg, router) error`) using go-chi             |
| `httpserver/middlewares`  | Middleware system (3-layer: config → handler wrapper → handler)             |
| `httpserver/render`      | Response rendering (`render.JSON`, `render.Error`, `render.Text`)          |
| `data`                   | PostgreSQL/sqlx database layer (context-threaded `*sqlx.DB`)               |
| `data/migrator`          | SQL migration engine (file-based + `go:embed` support)                     |
| `data/page`              | Pagination helpers                                                         |
| `cmd`                    | CLI commands via cobra (`serve`, `print-config`, data subcommands)         |
| `fxlog`                  | Structured logging (zerolog default, slog option)                          |
| `worker`                 | PostgreSQL-backed background job system                                    |
| `validate`               | Input validation helpers                                                   |
| `errutil`                | Error decoration (`WithCode`, `WithData`, `Wrap`)                          |
| `cache`                  | In-memory / Redis caching abstraction                                      |
| `blobstore`              | S3-compatible object storage                                               |
| `secret`                 | AES-256-GCM encryption (`Hide`/`Reveal`)                                   |
| `passwords`              | bcrypt password hashing                                                    |
| `mailer`                 | Postmark email integration                                                 |
| `clients/`               | Third-party clients (Coda, OpenAI, Typesense)                              |
| `fxtest`                 | Test utilities (`fxtest.Configure()` for test config)                      |
| `ctrlc`                  | Graceful CTRL-C signal handling                                            |
| `slices`                 | Slice utilities                                                            |

## Key Patterns

- **App Fragments**: Modular composition via `app.Build().Name("x").Controllers(...).Mount(child)`.
- **Controllers**: Implement `controllers.Interface` with `Mount(cfg *config.Source, router chi.Router) error`.
- **Middlewares**: `func(cfg *config.Source) func(http.Handler) http.Handler` (3-layer nesting).
- **Config Vars**: Declare `var X = config.StrDef("ENV_VAR", "default")`, read with `config.Get(cfg, X)`.
- **Database**: Context-threaded `*sqlx.DB`; use `data.Get`/`data.Select`/`data.Exec` or `data.Run(ctx, func(Scope) error)` for transactions.
- **Transactions**: `data.Scope` with auto commit/rollback; prefer `data.Run()` closure style.
- **Rendering**: `render.JSON(w, r, obj)`, `render.Error(w, r, status, err)`.

## Key Dependencies

- `go-chi/chi/v5` — HTTP routing
- `spf13/cobra` — CLI commands
- `jmoiron/sqlx` + `jackc/pgx/v5` — Database
- `rs/zerolog` — Logging
- `redis/go-redis/v9` — Redis
- `getsentry/sentry-go` — Error tracking
- `rs/cors` — CORS middleware
- `joho/godotenv` — .env file loading
- `stretchr/testify` — Testing

## Running

```sh
go run . serve           # Start HTTP server
go run . print-config    # Debug configuration
go run . data migrate    # Run database migrations
go run . data rollback   # Revert last migration
go run . data create-db  # Create database
go run . data psql       # Open psql shell
go test ./...            # Run tests
```

## Testing

Tests use `testify` for assertions and `fxtest.Configure()` for test config sources.
Test files exist in: `validate/`, `secret/`, `data/migrator/`, `data/dbname/`, `examples/envfiles/`.

## Releasing

Uses `platform` CLI with semver strategy (configured in `platform.toml`).
Only update `CHANGELOG.md` when releasing, not per-commit.

```sh
platform release --patch    # Increment patch version
platform release --minor    # Increment minor version
platform release --major    # Increment major version
platform release v1.2.3     # Explicit version name
platform release --force    # Release even if worktree is dirty
```

## Documentation

- `DOCS.md` — Comprehensive framework documentation (config, app fragments, controllers, middlewares, database, transactions, migrations, logging)
- `README.md` — Project overview, git subtree workflow, vanity server
- `examples/` — Reference implementations (todoapi, envfiles, workers)
