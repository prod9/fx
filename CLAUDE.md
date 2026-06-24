# FX — PRODIGY9 Go API Framework

Module: `fx.prodigy9.co` | Go 1.24 | Maintainer: Chakrit Wichian

FX is a minimalistic, modular Go API framework. It bundles well-integrated tools for
building APIs while letting engineers swap pieces in/out. Most projects should just
`go get fx.prodigy9.co` and import the packages they need. `git subtree` remains a
supported escape hatch for hacking on FX itself from a downstream app, but is no
longer the default install path.

## Philosophy

Full version: [`docs/spec/philosophy.md`](docs/spec/philosophy.md). Apply these as
design-intent tests for any non-trivial change:

1. **Modular by composition, not by config.** New cross-cutting capabilities ship as app
   fragments, not registries or hook systems.
2. **Thin wrappers over standard primitives, never replacements.** chi, sqlx, cobra,
   `net/http` stay reachable. Add ergonomics, not opacity.
3. **Config as decentralized declarations.** New tunables declare `config.*Def` next to
   their consumer. No central `Config` struct.
4. **Context as the carry-bag.** Ambient values (config, DB, request-scoped state) ride
   `context.Context`. Don't invent parallel passing mechanisms.
5. **Everything is optional.** New packages must work standalone. Cross-package coupling
   needs strong justification.
6. **Operational concerns are first-class.** Any new subsystem answers "how do I debug
   this in prod?" before it ships — usually a `cmd` subcommand or introspection route.
7. **Punt distributed-systems problems to infrastructure.** No service mesh, no retry
   framework, no pluggable transport. Document the infra assumption instead.
8. **Convention over options, but one layer deep.** One recommended path plus a small
   number of escape hatches. Reject pluggable-everything designs.

Minimalism here is a discipline against accidental complexity, not a goal in itself.

## Durable artifacts

`docs/{notes,decisions,spec}/` — sorted by permanence (impermanent / point-in-time /
current). Default to `notes/`. See `docs/README.md` and per-dir READMEs for picker
details.

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
| `app/audit`              | Append-only audit trail (`Record`/`Log`/`List`, `audit_events` migration) |
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
| `clients/`               | Third-party clients (OpenAI, Typesense)                                    |
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

## Git commit convention

Format: `<scope>: <Sentence-case summary>` — no trailing period, imperative mood. No
parentheses in the scope (no `feat(app):`, no `docs(TODO):`).

Scope is a single label, in rough order of preference:

- Package or area name: `app:`, `cmd:`, `data:`, `migrator:`, `examples:`, `docs:`,
  `platform:`, `root:`.
- Plain conventional type when no single package fits: `feat:`, `fix:`, `refactor:`,
  `chore:`.

Subject is Sentence case ("Add…", "Fix…", "Move…"), short, parenthetical clarifiers
allowed in the subject itself. No `fx:` prefix (legacy from the subtree era). No
Claude/co-author trailers — commits land under the maintainer's identity, plain.

## Releasing

See [`docs/spec/releasing.md`](docs/spec/releasing.md). Short version: clean tree, push
commits, update `CHANGELOG.md`, `./platform release --patch` (tags + pushes the tag).

## Documentation

- `docs/spec/` — Per-topic specs (philosophy, configuration, app-fragments,
  controllers, middlewares, database, migrations, logging, workers, mailer, errors,
  testing, utilities). Canonical reference.
- `docs/{decisions,notes}/` — Point-in-time rulings and impermanent notes (see
  `docs/README.md`).
- `docs/TODO.md` — Running follow-up list.
- `DOCS.md` — Now a thin index pointing into `docs/spec/`; kept for anyone who
  bookmarked the old path.
- `README.md` — Project overview, install, philosophy summary, vanity server.
- `examples/` — Reference implementations (todoapi, envfiles, workers, migrations).

## Load these skills

Default skills for this project (drives `ace.toml` `skills` filter):

- `ace`, `ace-*` — ACE workflow / session management
- `general-coding` — base coding workflow
- `go-coding` — Go conventions
- `prod9-fx` — this framework's own conventions and API reference
- `markdown-writing` — for DOCS.md / README.md / CHANGELOG.md edits
- `shell` — release/run scripts
- `rtk` — shell-output compaction
- `skill-creator` — school skill edits propagate from here
- `issue-creator` — ticket drafting
- `note-taker` — meeting/discussion capture
