# App Fragments

**Status:** accepted

Applications built on `fx` are expected to be somewhat modular. It borrows a little bit
from Django's apps concept. App fragments bundle together a set of related stuff into a
composable unit. This includes:

* Controllers
* Middlewares
* Commands
* Embedded migrations (aggregated up through `Mount`)

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
* `.Start()` builds up a cobra's root command from all the fragments and runs it.
* Embedded migrations on child fragments are picked up automatically — each fragment's
  `EmbedMigrations(fs)` registers with the migrator at `Start()` time.

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

* `go run . print-config` — Prints resolved configuration, useful for debugging.
* `go run . serve` — Starts HTTP server.
* `go run . data migrate` — Run all pending database migrations.
* `go run . data rollback` — Revert the last applied migration.

## Built-in App Fragments

### `settings.App`

Key-value settings stored in PostgreSQL with a REST API and config provider. Include it
by mounting the fragment:

```go
app.Build().
  Mount(settings.App).
  Start()
```

Provides CRUD functions: `settings.List()`, `settings.Get()`, `settings.Set()`,
`settings.Delete()`.

### `files.App` / `files.NewApp(client)`

S3-backed file management with presigned URL uploads, metadata stored in PostgreSQL,
and single/multi-file controllers. Uses the `blobstore` package for S3 operations.

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
