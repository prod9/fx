# FX Documentation

FX is short for framework. It provides a well-integrated set of tools for building Go APIs
while staying modular — individual components can be swapped in or out as needed.

## Configuration

The `config` package allows defining configuration via environment variables in a way that
you don't end up with one giant `struct` and each module can maintain its own list of
configurable things.

Most modules in `fx` accept either a `context.Context` (which is expected to carry a
`*config.Source` inside it) or a `*config.Source` directly.

Start by creating a configuration variable:

```go
var SessionTimeoutConfig = config.DurationDef("SESSION_TIMEOUT", 3*time.Minute)
```

Available definitions include:

* `config.Str` - Plain string
* `config.Int` - Integer
* `config.Int64` - 64-bit integer
* `config.URL` - URL (wraps `url.Parse`)
* `config.Bool` - Boolean (wraps `strconv.ParseBool`)
* `config.Duration` - Duration (wraps `time.ParseDuration`)

Everything should have a `*Def` version such as `config.StrDef` which takes a default
value when the environment variable is not set, or incorrectly set.

Then, where you need the actual config values, get from `config.Get`:

```go
func processSession(src *config.Source) {
  timeout := config.Get(src, SessionTimeoutConfig)

  // use timeout
}
```

Usually you'll get the `*config.Source` from `request.Context()` like so:

```go
func Index(resp http.ResponseWriter, req *http.Request) {
  src := config.FromContext(req.Context())
  processSession(src)
}
```

If you need to pass the config around, we usually bundle it into whatever `Context` we
have lying around:

```go
ctx := context.Background()
cfg := config.Configure()

ctx = config.NewContext(ctx, cfg)
performWork(ctx)
```

You can use `config.Configure()` to just straight up get a new source to work with when no
context is available.

You can also `config.Set` or `config.SetDefault` to override values where needed, such as
in tests or when defaults in some modules are not what you want.

```go
// permanent overrides
config.Set(data.DatabaseMaxIdleConfig, 2)

// just the defaults, if not specified
config.SetDefault(data.DatabaseURLConfig, "postgres:///mydb")
```

Configurations are read from the following list, in order, with higher ones overriding
lower ones:

* `config.Set` - Values overridden at runtime.
* Actual Environment variables
* `.env.local` and `.env` files — searched from the current directory upward, stopping at
  the nearest `.git` directory. At each level, `.env.local` overrides `.env`, and closer
  files override those further up. Useful for monorepo setups with shared envs.
* `config.SetDefault` - Default values overridden at runtime.
* `config.*Def` - Default values set on definition.
* Go defaults (e.g. `0` for int, `""` for string, etc.)

## App Fragments

Applications built on `fx` are expected to be somewhat modular. It borrows a little bit
from Django's apps concept. Mainly, app fragments bundle together a set of related stuff
into a composable unit. This includes:

* Controllers
* Middlewares
* Commands

Create sub-apps by calling `app.Build` and just using the available methods:

```go
// in file auth/auth.go
var App = app.Build().
  Name("auth").
  Controllers(
    &SessionController{},
    &UserController{},
  ).
  Commands(
    CreateAdminCmd,
  )

// in file todo/todo.go
var App = app.Build().
  Name("todo").
  Controllers(
    &TodoCtr{},
  )
```

Then in your `main.go` file, you can compose them together like so:

```go
package main

import (
  "yourapp/auth"
  "yourapp/todo"

  "fx.prodigy9.co/app"
)

func main() {
  err := app.Build().
    Name("my todo app").
    AddDefaults().
    Mount(auth.App).
    Mount(todo.App).
    Start()

  if err != nil {
    log.Fatalln(err)
  }
}
```

Couple things to note:

* Controllers wrap `github.com/go-chi/chi` routers.
* Commands are `github.com/spf13/cobra` commands.
* `.Start()` builds up a cobra's root command from all the fragments and run it.

Once composed, your `main` will become a CLI application with a few useful commands:

```
PRODIGY9 FX Application

Usage:
  app [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  data         Work with databases
  help         Help about any command
  print-config Prints current effective configuration.
  serve        Starts an HTTP server.

Flags:
  -h, --help   help for app

Use "app [command] --help" for more information about a command.
```

Notable ones are:

* `go run . print-config` - Prints resolved configuration, useful for debugging.
* `go run . serve` - Starts HTTP server.
* `go run . data migrate` - Run all pending database migrations.
* `go run . data rollback` - Revert the last applied migration.

## Controllers

Controllers are just an interface with one method:

```go
// package controllers
type Interface interface {
  Mount(cfg *config.Source, router chi.Router) error
}
```

Normally we define controllers in a concern-specific package such as `todos` or `auth`.

```go
// todos/todo_ctr.go
type TodoCtr struct{}

var _ controllers.Interface = (*TodoCtr)(nil)

func (c *TodoCtr) Mount(cfg *config.Source, r chi.Router) error {
  r.Route("/todos", func(r chi.Router) {
    r.Get("/", c.List)
    r.Post("/", c.Create)
    r.Get("/{id}", c.Get)
    r.Put("/{id}", c.Update)
    r.Delete("/{id}", c.Delete)
  })
  return nil
}

func (c *TodoCtr) List(w http.ResponseWriter, r *http.Request) {
  // list todos
}

// ...
```

Handlers are usually implemented as a simple branching call to a service method and
usually ends with a call to the `render` package:

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

The `render` package has the following methods:

* `render.Text` - Renders plain text
* `render.JSON` - Renders JSON
* `render.Redirect` - Redirects (307) to a URL
* `render.FileTransfer` - Transfers a file (force a download)
* `render.Error` - Renders error JSON

## Built-in Components

### Built-in Controllers

The `httpserver/controllers` package includes several ready-to-use controllers:

* `Home{}` — Mounts `GET /` returning `{"time": "..."}` with the current server time.
  Useful for deployment testing and basic health checks.
* `Debug{}` — Mounts `GET /__panic` which triggers a test panic. Useful for verifying
  panic recovery middleware and error reporting (e.g. Sentry).
* `StaticJSON(path, obj)` — Creates a controller that serves a static JSON object at the
  given path via `GET`.
* `FromFunc(path, handlerFunc)` — Wraps an `http.HandlerFunc` as a controller mounted at
  the given path.
* `FromHandler(path, handler)` — Wraps an `http.Handler` as a controller mounted at the
  given path.

### Built-in App Fragments

#### `settings.App`

Key-value settings stored in PostgreSQL with a REST API and config provider. Include it
by mounting the fragment:

```go
app.Build().
  Mount(settings.App).
  Start()
```

Provides CRUD functions: `settings.List()`, `settings.Get()`, `settings.Set()`,
`settings.Delete()`.

#### `files.App` / `files.NewApp(client)`

S3-backed file management with presigned URL uploads, metadata stored in PostgreSQL, and
single/multi-file controllers. Uses the `blobstore` package for S3 operations.

```go
import "fx.prodigy9.co/app/files"

// Mount the fragment for migrations (uses global blobstore)
app.Build().
  Mount(files.App).
  Start()

// Or with a specific blobstore client
client := blobstore.NewClient(cfg)
app.Build().
  Mount(files.NewApp(client)).
  Start()
```

**Defining file kinds** — each kind describes a type of file attachment:

```go
var userAvatar = files.Kind{
  Name: "user-avatar", Multiple: false,
  OwnerType: "user", ContentTypes: files.ImageTypes,
}

var projectDocs = files.Kind{
  Name: "project-doc", Multiple: true,
  OwnerType: "project", ContentTypes: []string{"application/pdf", "image/png"},
}
```

**Mounting file controllers** inline within your own controllers:

```go
func (c *UserCtr) Mount(cfg *config.Source, r chi.Router) error {
  r.Route("/users/{id}/avatar", func(r chi.Router) {
    files.Controller(userAvatar,
      files.WithMode(files.ModeReadOnly),
      files.WithLinkAge(5*time.Minute),
    ).Mount(cfg, r)
  })
  return nil
}
```

`Kind.Multiple` controls which controller type is used:
* `false` → single-file controller (`GET /`, `GET /meta`, `POST /`, `DELETE /`)
* `true` → multi-file controller (`GET /`, `GET /{fileID}`, `GET /{fileID}/meta`,
  `POST /`, `DELETE /{fileID}`)

**Controller options:**

* `WithKind(kind)` — Override the file kind.
* `WithMode(mode)` — `ModeReadOnly` or `ModeReadWrite` (default).
* `WithOwnerIDFunc(func(*http.Request) int64)` — Custom owner ID extraction (default:
  reads `{id}` URL param).
* `WithClient(*blobstore.Client)` — Use a specific blobstore client instead of the
  global default.
* `WithLinkAge(duration)` — Per-controller presigned URL expiry override.

**Configuration:**

* `FILE_LINK_AGE` — App-wide default presigned URL expiry (default: `1m`). Override
  per-controller with `WithLinkAge`.

## Middlewares

Middlewares in `fx` attempts 2 things: Conform to go's standard `http.Handler` signature
and be easy to use with `github.com/go-chi/chi`.

They are a bit complicated, but not hard to understand. There are 3 layers of `func`
nesting:

1. First layer takes in a Configuration to configure the middleware.
2. Second layer is a classic `http.Handler` wrapper function (as required by `go-chi`).
3. Third layer is the actual middleware code.

```go
// use this in your Mount() calls:
func RequirePermission(cfg *config.Source, permission string) func(http.Handler) http.Handler {

  // go-chi js-style next wrapper
  return func(next http.Handler) http.Handler {

    // actual middleware handler
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

      if !checkPermission(r.Context(), permission) {
        render.Error(w, r, 403, httperrors.ErrForbidden)
        return
      }
      next.ServeHTTP(w, r)

    })
  }
}
```

Points to note:

* The first layer is expected to be called multiple times throughout your application to
  create different instances of the middleware. You will need to be careful with handling
  any global state, if needed.

* If you are importing a 3rd-party middleware, you usually only need the 2nd layer.
  For example, to add sentry:

  ```go
  import "github.com/getsentry/sentry-go"

  // get dsn using config package
  func Sentry(cfg *config.Source) func(http.Handler) http.Handler {
    dsn := config.Get(cfg, SentryDSNConfig) // optional, if you have custom DSN code

    // lets sentryhttp do the middleware wrapping
    return func(next http.Handler) http.Handler {
      sentry.Init(sentry.ClientOptions{Dsn: dsn})

      return sentryhttp.New(sentryhttp.Options{
        Repanic: true,
        WaitForDelivery: false,
      }).Handle(next)
    }
  }
  ```

Some default middlewares are included when `.AddDefaults()` is called:

* `middlewares.Configure` - Injects `*config.Source` into the request context.
* `middlewares.LogRequests` - Capture metrics with `github.com/felixge/httpsnoop` and log requests/responses.
* `middlewares.CORSAllowAll()` - Wraps `github.com/rs/cors` to allow all CORS requests.
* `middlewares.AddDataContext` - Adds a `*sqlx.DB` to the request context.

Some other non-default middlewares includes:

* `middlewares.CheckMigrations` - If `MUST_MIGRATE=1` is set, all routes will returns
  errors until all known migrations are applied.
* `middlewares.DebugRequest` - Prints out incoming request body if `DEBUG_REQUEST=1` is
  set or prints out the outgoing response body if `DEBUG_RESPONSE=1` is set.
* `middlewares.Sentry` - Installs sentry error reporting handler with DSN set in
  `API_SENTRY_DSN` environment variable.

Use go-chi's `Route` and `Group` to create sub-routes and apply middlewares selectively.

```go
r.Route("/todos", func(r chi.Router) {
  r.Get("/", c.List) // public todo list

  // modification requires authentication
  r.Group(func(r chi.Router) {
    r.Use(auth.RequirePermission(cfg, "manage_own_todos"))
    r.Post("/", c.Create)
  })

  // ...
})
```

## Working with database

The `data` package is completely optional. It mainly wraps `github.com/jmoiron/sqlx` to
make it a bit easier to control transaction scopes, properly handle rollbacks on errors
etc.

Configure with the `DATABASE_URL` environment variable, usually just adds an entry in
`.env`:

```sh
DATABASE_URL=postgres:///mynewdb?sslmode=disable
```

Connects to the db with:

```go
// using DATABASE_URL env var
db, err := data.Connect(cfg)

// override with config.Set
config.Set(data.DatabaseURLConfig, "postgres:///mydb")
db, err := data.Connect(cfg)

// manually adding data middleware
err := app.Build().
	Name("my todo app").
	Middlewares(middlewares.AddDataContext()).
	// ...
	Start()
```

HA setups, if needed, is assumed to be handled outside the application at the
infrastructure level with something like `pgbouncer`.

Once setup, there are basic methods to interact with the database:

* `data.Get` - Executes a query and returns a single row.
* `data.Select` - Executes a query and returns multiple rows.
* `data.Exec` - Executes a query with no return value.
* `data.Run` - Runs a function inside a transaction (see Transactions section below).
* `data.Prepare` - Prepares a statement and returns a `*sqlx.Stmt` object.

Example:

```go
func GetTodoByID(ctx context.Context, id string) (*Todo, error) {
  const sql = "SELECT * FROM todos WHERE id = $1"

  todo := &Todo{}
  if err := data.Get(ctx, todo, sql, id); err != nil {
    return nil, err
  } else {
    return todo, nil
  }
}
```

Note that the `*sqlx.DB` is assumed to already be in the context. If not, you can use
`data.NewContext` like with the `config` package to pass it around:

```go
cfg := config.Configure()
db := data.MustConnect(cfg)

ctx := context.Background()
ctx = data.NewContext(ctx, db)
```

The methods usually takes the following parameters, in order:

* `ctx` - Context, usually from the request.
* `dest` - Destination object to scan into. Arrays for `Select`, structs for `Get`.
* `sql` - SQL query to execute.
* `args` - Arguments to pass to the query.

There's an unfinished version of support for SQL Generators like `go-jet` with `GetSQL`
and similar. They should work, but largely untested, and very alpha.

## Transactions

Transactions are controlled through `data.Scope`. A scope wraps a `*sqlx.Tx` and
automatically issues `COMMIT` or `ROLLBACK` based on the returned error.

### `data.Run` (recommended)

The simplest way to run a transaction. The scope is created and ended automatically:

```go
func GetUserByID(ctx context.Context, out any, id int64) error {
	return data.Run(ctx, func(scope data.Scope) error {
		err := scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
		if err != nil {
			return err
		}

		// pass scope.Context() to propagate the transaction to nested calls
		var profile *UserProfile
		if err = GetUserProfile(scope.Context(), profile, id); err != nil {
			return err
		}

		user.Profile = profile
		return nil
	})
}
```

### `data.NewScopeErr` (flat style)

If you prefer a flat function body over a closure, `data.NewScopeErr` returns a cancel
function to `defer`. It takes a pointer to the return error so it can decide whether to
commit or rollback:

```go
func GetUserByID(ctx context.Context, out any, id int64) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()

	// watch out for the scope of the `err` variable
	err = scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	var profile *UserProfile
	if err = GetUserProfile(scope.Context(), profile, id); err != nil {
		return err
	}

	user.Profile = profile
	return nil
}
```

### Manual scope management

For full control, create a scope with `data.NewScope` and call `scope.End(&err)` in a
`defer`:

```go
func GetUserByID(ctx context.Context, out any, id int64) (err error) {
	var scope data.Scope
	scope, err = data.NewScope(ctx, nil)
	if err != nil {
		return err
	}
	defer scope.End(&err)

	err = scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	var profile *UserProfile
	if err = GetUserProfile(scope.Context(), profile, id); err != nil {
		return err
	}

	user.Profile = profile
	return nil
}
```

### Notes

* A pointer to the return error is passed so the scope can automatically rollback on
  error.
* `data.Scope` is an interface wrapping `*sqlx.Tx` — it provides `Get`, `Select`, `Exec`,
  and `Prepare` methods mirroring the top-level `data` functions.

## Database Migrations

A built-in migration engine is provided in `data/migrator`. To use this, adds data
commands to the application:

```go
app := app.Build().
  Name("my todo app").
  // or just .AddDefaults() which already includes data commands
  Commands(data.Cmd).
```

The following commands become available:

* `go run . data migrate` - Runs all migrations.
* `go run . data new-migration (name) [subdir]` - Creates new up+down migration files.

Migrations are written as normal SQL files. Usually they contains `CREATE TABLE` for the
up migration and `DROP TABLE` for the down migration.

During production deployment, migrations can be collected and embedded into the
application itself using `go:embed` directive for easy distribution:

```go
//go:embed **/*.sql
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

The migrator will automatically look for embedded sources. Otherwise it will look at the
current folder, and its parents for the files. Uses the `data list-migrations` command to
check:

```sh
$ go run ./api data list-migrations

api/auth/202312281812_create_users_and_sessions.up.sql
api/listing/202504011719_create_listing.up.sql
api/files/202504041033_create_files.up.sql
```

Other commands include:

* `go run . data collect-migrations (outdir)` - Collect migration files into a single directory.
* `go run . data create-db` - Creates database specified in the config.
* `go run . data list-migrations` - List all detected migration files.
* `go run . data migrate` - Runs all detected migration scripts.
* `go run . data new-migration (name) [subdir]` - Creates new up+down migration files.
* `go run . data psql` - Starts a psql shell connecting to the configured database.
* `go run . data recover-migrations [output-dir]` - Export migration cache from database to files.
* `go run . data resync-migrations` - Update database migration cache to match program files.
* `go run . data rollback` - Revert one previously ran migration.

### Scripting and CI

Data commands use the `cmd/prompts` package for interactive input. To run commands
non-interactively (e.g. in CI/CD pipelines or scripts), set `CI=1`. In CI mode, all
required inputs must be provided as positional arguments — if any are missing, the command
will exit with an error instead of prompting.

Additionally, set `ALWAYS_YES=1` to automatically confirm all yes/no prompts.

## Logging

Logging in fx is done using the `fxlog` subpackage. It is pre-configured to output
pretty structured logs by default. It has 3 basic functions, mirroring standard library
log package:

* `fxlog.Log` - Logs general messages.
* `fxlog.Error` - Logs errors.
* `fxlog.Fatal` - Logs fatal errors and exits the application.

For `Error` and `Fatal`, there are also `Errorf` and `Fatalf` variants that simply calls
`fmt.Errorf` to format the message for you.

Log output is set to the default `zerolog` logger by default. There are a few ways to
override the output and customize logging behavior:

1. Switch the sink: set `LOG_SINK=slog` to redirect FX log outputs to `log/slog` default
   logger and configure `log/slog` normally as you would do in any other Go application.

2. Set a custom `log/slog` or `zerolog` logger (initialized outside of fx) by using
   `SetSink`:

```go
// zerolog
zl := zerolog.New()
fxlog.SetSink(fxlog.NewZerlogSink(zl))

// slog
sl := slog.New(slog.NewTextHandler(os.Stderr, nil))
fxlog.SetSink(fxlog.NewSlogSink(sl))
```

3. Create your own `fxlog.Sink` implementation for maximum customization or if you wish
   to use a different logging library that's not provided out of the box:

```go
mysink := NewCustomSink()
fxlog.SetSink(mysink)
```

## Testing

The `fxtest` package provides helpers for writing tests against `fx` components.

* `fxtest.Configure()` — Returns a `*config.Source` initialized for testing (reads `.env`
  files, applies defaults).
* `fxtest.ConnectTestDatabase(t)` — Creates an isolated test database and returns a
  `context.Context` carrying both the config source and `*sqlx.DB`. The database is
  automatically dropped when the test completes, unless `FXTEST_CLEANUP=no` is set.

```go
func TestSomething(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	// ctx carries *config.Source and *sqlx.DB — pass to data functions directly
	err := data.Exec(ctx, "INSERT INTO todos (title) VALUES ($1)", "test")
	require.NoError(t, err)
}
```

## Error Utilities

The `errutil` package provides helpers for decorating and collecting errors.

* `errutil.Wrap(name, &err)` — Intended for use in a `defer`. Prefixes the error with
  `name` if `*err` is non-nil.
* `errutil.WithCode(err, code)` — Attaches a string error code to the error (useful for
  API error responses).
* `errutil.WithData(err, data)` — Attaches arbitrary context data to the error.
* `errutil.NewCoded(code, msg, data)` — Creates a new error with code, message, and data.
* `errutil.Decorate(err)` — Wraps an error in a `decoratedErr` for JSON serialization.
* `errutil.Aggregate[T](slice, func)` — Runs a function on each element in parallel,
  collecting all errors into a single aggregated error.
* `errutil.AggregateWithTags[T](slice, func)` — Like `Aggregate` but each error is tagged
  with a label for identification.

## Background Workers

The `worker` package provides a PostgreSQL-backed background job system.

### Setup

Register job types and start the worker:

```go
worker := worker.New(cfg, &SendEmailJob{}, &CleanupJob{})
worker.Start() // blocks, polling for jobs
worker.Stop()  // graceful shutdown
```

### Job Interface

Jobs implement the `worker.Interface`:

```go
type Interface interface {
	Name() string           // unique job name, used as DB key
	Run(ctx context.Context) error
}
```

### Scheduling

```go
worker.ScheduleNow(ctx, &SendEmailJob{To: "user@example.com"})
worker.ScheduleIn(ctx, &CleanupJob{}, 30*time.Minute)
worker.ScheduleAt(ctx, &ReportJob{}, tomorrow)

// schedule only if a pending job with the same name doesn't already exist
worker.ScheduleNowIfNotExists(ctx, &DailyDigestJob{})
```

### Configuration

* `WORKER_POLL` — Polling interval (default: `1m`).

## Mailer

The `mailer` package sends transactional emails via Postmark.

```go
err := mailer.Send(cfg, &mailer.Mail{
	From:     "noreply@example.com",
	To:       []string{"user@example.com"},
	Subject:  "Welcome!",
	HTMLBody: "<h1>Hello</h1>",
	TextBody: "Hello",
})
```

### Configuration

* `POSTMARK_TOKEN` — Postmark server API token.

## Other Packages

* `fx.prodigy9.co/blobstore` — S3-compatible object storage client. Used by the `files`
  package for presigned URL uploads and downloads.
* `fx.prodigy9.co/cache` — In-memory and Redis caching with a unified interface.
* `fx.prodigy9.co/cmd/prompts` — Interactive TUI prompts for CLI commands (text input,
  list selection, yes/no confirmation). Inputs can be provided as positional args for
  scripting. Set `CI=1` for non-interactive mode, `ALWAYS_YES=1` to auto-confirm.
* `fx.prodigy9.co/ctrlc` — Graceful CTRL-C / SIGINT signal handler.
* `fx.prodigy9.co/passwords` — BCrypt password hashing.
* `fx.prodigy9.co/secret` — AES-256-GCM encryption for passing secrets safely (`Hide` /
  `Reveal`).
* `fx.prodigy9.co/validate` — Input validation helpers.

