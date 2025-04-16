# FX Documentation

FX is short for framework. It tries not to become a full-fleged framework, but instead a
convenience set of tools for putting together most of things that are needed to build a
proper functioning webapp. Add more libs as needed, etc.

## Configuration

The `config` package allows defining configuration via environment variables in a way that
you don't end up with one giant `struct` and each module can maintain its own list of
configurable things.

Most of the modules in `fx` will either accept a `context.Context` assuming it contains,
or requiring that you pass a `*config.Source` directly.

Start by creating a configuration variable:

```go
var SessionTimeoutConfig = config.DurationDef("SESSION_TIMEOUT", 3*time.Minute)
```

Available definitions include (as of 0.7):

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
* `.env.local` - Local overrides (should be ignored by git)
* `.env` - Development values (committed to git)
* `../.env.local` - Local overrides, for monorepo setups with shared envs (should be
  ignored by git)
* `../.env` - Development values, for monorepo setups with shared envs (committed to git)
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
  test-email   Sends a test email to check if SMTP configuration works

Flags:
  -h, --help   help for app

Use "app [command] --help" for more information about a command.
```

Notable ones are:

* `go run . print-config` - Prints resolved configuration, useful for debugging.
* `go run . serve` - Starts HTTP server.

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

The `render` package has the following methods as of v0.7:

* `render.Text` - Renders plain text
* `render.JSON` - Renders JSON
* `render.Redirect` - Redirects (307) to a URL
* `render.FileTransfer` - Transfers a file (force a download)
* `render.Error` - Renders error JSON

## Middlewares

Middlewares in `fx` attempts 2 things: Conform to go's standard `http.Handler` signature
and easy to use with `github.com/go-chi/chi`.

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
  create different intances of the middleware. You will need to be careful with handling
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
* `middlewares.LogRequests` - Capture metrics with `httpsnoop` and log requests/responses.
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
app := app.Build().
  Name("my todo app").
  Middlewares(middlewares.AddDataContext()).

  //...

  Start()
```

HA setups, if needed, is assumed to be handled outside the application at the
infrastructure level with something like `pgbouncer`.

Once setup, there are basic methods to interact with the database:

* `data.Exec` - Executes a query with no return value.
* `data.Get` - Executes a query and returns a single row.
* `data.Select` - Executes a query and returns multiple rows.
* `data.Prepare` - Prepares a statement and returns a `*sqlx.Stmt` object.

Example:

```go
func GetTodoByID(ctx context.Context, id string) (*Todo, error) {
  const sql = "SELECT * FROM todos WHERE id = $1"

  todo := &Todo{}
  if err := data.Get(db, todo, sql, id) ; err != nil {
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

Transactions are mostly controlled through the use of `data.Scope`. Create and `.End()` a
scope properly and transactions should be automatically `COMMIT`-ed or `ROLLBACK`-ed.

Originally, these used to be quite manual, posting here for reference:

```go
func GetUserByID(ctx context.Context, out any, id int64) (err error) {
  var scope data.Scope
  scope, err = data.NewScope(ctx, nil)
  if err != nil {
    return err
  }
  defer scope.End(&err)

  // instead of data.Get, use scope.Get
  err = scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
  if err != nil {
    return err
  }

  // instead of request.Context(), pass scope.Context()
  var profile *UserProfile
  if err = GetUserProfile(scope.Context(), profile, id); err != nil {
    return err
  } else {
    user.Profile = profile
    return nil
  }
}
```

In recent versions, there are 2 ways, both a little bit more ergonomic, to use scopes:
You can use `data.Run` to use closure-scoping.

```go
func GetUserByID(ctx context.Context, out any, id int64) (err error) {

  // scope.End is automatically handled
  return data.Run(ctx, func(scope *data.Scope) error {

    err := scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
    if err != nil {
      return err
    }

    // instead of request.Context(), pass scope.Context()
    var profile *UserProfile
    if err = GetUserProfile(scope.Context(), profile, id); err != nil {
      return err
    } else {
      user.Profile = profile
      return nil
    }
  })
}
```

Or you can use `data.NewScopeErr` which gives you a `context.CancelFunc` for you to
`defer`:

```go
func GetUserByID(ctx context.Context, out any, id int64) (err error) {
  scope, cancel, err := data.NewScopeErr(ctx, &err)
  defer cancel()

  err = scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
  if err != nil {
    return err
  }

  // instead of request.Context(), pass scope.Context()
  var profile *UserProfile
  if err = GetUserProfile(scope.Context(), profile, id); err != nil {
    return err
  } else {
    user.Profile = profile
    return nil
  }
}
```

Couple of points to note:

* Pointer to returning error is passed so that if an error is returned the scope will
  automatically rollback.
* Scope actually just wraps management of `*sqlx.Tx` which are actually just `*sql.Tx`
  underneath it. So you can use it as a normal transaction if you really want to.

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
* `go run . data new-migration` - Creates new up+down migration files.

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
    EmbedMigrations(sqlMigrations).

    // ...

    Start()

  if err != nil {
    log.Fatalln(err)
  }
}
```

The migrator will automatically looks for embedded sources. Otherwise it will look at the
current folder, and its parents for the files. Uses the `data list-migrations` command to
check:

```sh
$ go run ./api data list-migrations

api/auth/202312281812_create_users_and_sessions.up.sql
api/listing/202504011719_create_listing.up.sql
api/files/202504041033_create_files.up.sql
```

Other commands include:

* `go run . data create-db` - Creates database specified in the config.
* `go run . data list-migrations` - List all detected migration files.
* `go run . data migrate` - Runs all detected migration scripts.
* `go run . data psql` - Starts a psql shell connecting to the configured database.
* `go run . data rollback` - Revert one previously ran migration.

Sets `ALWAYS_YES=1` to skip confirmation prompts. This is useful for CI/CD pipelines.

## Misc

Couple of other useful packages (docs tbd later):

* `fx.prodigy9.co/blobstore` - Store stuff on S3-compatible storage.
* `fx.prodigy9.co/cache` - Dumb memory/redis cache.
* `fx.prodigy9.co/cmd/prompts` - Quick interactive TUI prompts. (enter input, yesno etc.)
* `fx.prodigy9.co/ctrlc` - CTRL-C signal handler. (Might be better to just wrap tini)
* `fx.prodigy9.co/passwords` - BCrypt-hash passwords.
* `fx.prodigy9.co/secret` - Pass secret message around the internet safely.
* `fx.prodigy9.co/validate` - Validations.
* `fx.prodigy9.co/worker` - PSQL-backed Background worker (alpha, it works, but not fully tested).
