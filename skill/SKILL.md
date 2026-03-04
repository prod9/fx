---
name: prod9-fx
description: >
  Coding conventions and API reference for the prod9/fx Go web framework
  (fx.prodigy9.co). Load for any Go project using prod9/fx — covers app
  fragments, controllers, middlewares, config, data/sqlx, migrations,
  transactions, and logging.
  TRIGGER when: go.mod contains fx.prodigy9.co, or code imports fx.prodigy9.co/*.
---

# prod9/fx Framework

Go web framework providing composable app fragments with chi router, cobra CLI,
and sqlx data layer. Import path: `fx.prodigy9.co`.

## Reference Files

Load these on demand based on the task:

- **[references/database.md](references/database.md)** — Read when working with `data`
  package: queries (`data.Get`/`Select`/`Exec`), transactions (`data.Run`,
  `data.NewScopeErr`, `data.NewScope`), and migrations (`data/migrator`, CLI commands).
- **[references/middlewares-and-extras.md](references/middlewares-and-extras.md)** — Read
  when writing middlewares (3-layer pattern), configuring logging (`fxlog`), writing tests
  (`fxtest`), or using `errutil`, `worker`, or `mailer` packages.

## App Structure

Build applications by composing fragments. Each fragment bundles controllers,
middlewares, and CLI commands.

```go
// sub-app fragment (e.g. auth/auth.go)
var App = app.Build().
	Name("auth").
	Controllers(&SessionCtr{}, &UserCtr{}).
	Commands(CreateAdminCmd)

// main.go — compose fragments
func main() {
	err := app.Build().
		Name("myapp").
		AddDefaults().
		Mount(auth.App).
		Mount(todo.App).
		Start()
	if err != nil {
		log.Fatalln(err)
	}
}
```

`AddDefaults()` includes: config injection, request logging (httpsnoop), CORS
allow-all, data context middleware, and the `data` + `print-config` CLI commands.

Built-in CLI commands after composition:

- `go run . serve` — start HTTP server
- `go run . print-config` — print resolved config
- `go run . data migrate` — run pending migrations
- `go run . data rollback` — revert last migration
- `go run . data create-db` — create database
- `go run . data psql` — psql shell

Set `ALWAYS_YES=1` to skip confirmation prompts (CI/CD).

## Configuration

Define config vars anywhere — each module owns its config. No giant central struct.

```go
var TimeoutCfg = config.DurationDef("SESSION_TIMEOUT", 3*time.Minute)

timeout := config.Get(src, TimeoutCfg)           // from *config.Source
src := config.FromContext(req.Context())           // from request context
cfg := config.Configure()                          // standalone source
ctx = config.NewContext(ctx, cfg)                   // bundle into context

config.Set(data.DatabaseMaxIdleConfig, 2)           // runtime override
config.SetDefault(data.DatabaseURLConfig, "postgres:///mydb")
```

Types: `Str`, `Int`, `Int64`, `URL`, `Bool`, `Duration`. Each has a `*Def` variant.

Precedence (highest first): `config.Set` > env vars > `.env.local`/`.env` (searched
upward to nearest `.git`) > `config.SetDefault` > `*Def` defaults > Go zero values.

## Controllers

Interface with one method. Uses chi router.

```go
type TodoCtr struct{}

var _ controllers.Interface = (*TodoCtr)(nil)

func (c *TodoCtr) Mount(cfg *config.Source, r chi.Router) error {
	r.Route("/todos", func(r chi.Router) {
		r.Get("/", c.List)
		r.Post("/", c.Create)
		r.Get("/{id}", c.Get)
	})
	return nil
}
```

Handler pattern — branch on service result, end with render:

```go
func (c *TodoCtr) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if todo, err := GetTodoByID(r.Context(), id); err != nil {
		if data.IsNoRows(err) {
			render.Error(w, r, 404, httperrors.ErrNotFound)
		} else {
			render.Error(w, r, 500, err)
		}
	} else {
		render.JSON(w, r, todo)
	}
}
```

Render methods: `render.Text`, `render.JSON`, `render.Redirect` (307),
`render.FileTransfer`, `render.Error`.

## Built-in Components

**Controllers** (`httpserver/controllers`):

- `Home{}` — `GET /` returning `{"time": "..."}` (health check)
- `Debug{}` — `GET /__panic` (test panic recovery)
- `StaticJSON(path, obj)` — serve static JSON at path
- `FromFunc(path, handlerFunc)` / `FromHandler(path, handler)` — wrap handlers

**App fragments:**

- `settings.App` — key-value settings in PostgreSQL with REST API
- `files.App` / `files.NewApp(client)` — S3-backed file management with presigned URLs

## Package Map

| Package                  | Purpose                                                    |
|--------------------------|-------------------------------------------------------------|
| `app`                    | Application builder, fragment composition, lifecycle        |
| `app/files`              | S3-backed file management (presigned URLs, PostgreSQL meta) |
| `config`                 | Type-safe env var config (`Var[T]`, `.env`/`.env.local`)    |
| `httpserver`             | HTTP server startup                                         |
| `httpserver/controllers` | Controller interface using go-chi                           |
| `httpserver/middlewares`  | Middleware system (3-layer: config → wrapper → handler)      |
| `httpserver/render`      | Response rendering (`JSON`, `Error`, `Text`, etc.)          |
| `data`                   | PostgreSQL/sqlx database layer (context-threaded)           |
| `data/migrator`          | SQL migration engine (file-based + `go:embed`)              |
| `data/page`              | Pagination helpers                                          |
| `cmd`                    | CLI commands via cobra                                      |
| `fxlog`                  | Structured logging (zerolog default, slog option)           |
| `worker`                 | PostgreSQL-backed background jobs                           |
| `validate`               | Input validation helpers                                    |
| `errutil`                | Error decoration (`WithCode`, `WithData`, `Wrap`)           |
| `cache`                  | In-memory / Redis caching                                   |
| `blobstore`              | S3-compatible object storage                                |
| `secret`                 | AES-256-GCM encryption (`Hide`/`Reveal`)                    |
| `passwords`              | bcrypt password hashing                                     |
| `mailer`                 | Postmark email integration                                  |
| `fxtest`                 | Test utilities (`Configure()`, `ConnectTestDatabase(t)`)    |
| `ctrlc`                  | Graceful CTRL-C signal handling                             |
