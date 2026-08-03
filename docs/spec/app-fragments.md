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

Key-value settings stored in PostgreSQL with a config provider and an optional REST API.
Mount the fragment for the data layer; the table self-initializes on first access (and via
the embedded migration on deploy), so no consumer migration is required:

```go
app.Build().
  Mount(settings.App).
  Start()
```

Read/write from Go: `settings.List(ctx)`, `settings.Get(ctx, key, fallback)` (returns the
fallback when the key is absent — never an error), the `settings.Upsert{Key, Value}` action
(`Execute(ctx, out)`), and `settings.Delete(ctx, key)`.

The REST controller is **not** auto-mounted — settings data would otherwise be exposed by
the mere act of mounting the fragment. Mount it yourself inside a route group you guard, so
auth / RBAC / IP-allowlisting is a requirement, not a convention:

```go
r.Route("/admin", func(r chi.Router) {
  r.Use(adminAuth)                 // your guard
  settings.MountRoutes(cfg, r)     // wires GET/POST/DELETE /settings under it
})
```

### `files.App`

S3-backed file management with presigned URL uploads, metadata stored in PostgreSQL,
and single/multi-file controllers. The store is reached through the global `blobstore`
funcs (config-driven via `STORAGE_URL`) — the fragment is self-contained, so mounting it
takes no client:

```go
import "fx.prodigy9.co/app/files"

app.Build().
  Mount(files.App).
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
  MaxSize: 20 << 20, // 20 MiB cap; 0 ⇒ 256 MiB default, -1 ⇒ unlimited
}
```

**Mounting file controllers** inline within your own controllers:

```go
func (c *UserCtr) Mount(cfg *config.Source, r chi.Router) error {
  r.Route("/users/{id}/avatar", func(r chi.Router) {
    files.Controller(userAvatar,
      files.WithMode(files.ModeReadWrite),
    ).Mount(cfg, r)
  })
  return nil
}
```

`Kind.Multiple` controls which controller type is used:

* `false` → single-file controller (`GET /`, `GET /meta`, `POST /`, `DELETE /`)
* `true` → multi-file controller (`GET /`, `GET /{fileID}`, `GET /{fileID}/meta`,
  `POST /`, `DELETE /{fileID}`)

**Controller options** (kind is the positional arg, not an option):

* `WithMode(mode)` — `ModeReadOnly` (default) or `ModeReadWrite`. Writes are opt-in:
  read/write endpoints only mount when `ModeReadWrite` is set.
* `WithOwnerIDFunc(func(*http.Request) int64)` — Custom owner ID extraction (default:
  reads `{id}` URL param).

The presigned URL TTL is a fixed 1-minute package constant — access is gated by the route
that embeds the controller, so the TTL is policy, not a per-request knob.

**Cleanup worker** — `files.App` registers a `files-cleanup` worker that prunes abandoned
uploads: rows whose object never landed and that are older than 24h. It is driven off the
`files` table — each candidate row's object is probed directly — so it never enumerates or
deletes from the bucket; only `files` rows are removed. It is inert until seeded: call
`worker.ScheduleNowIfNotExists(ctx, files.CleanupJob)` once at startup (with a running
`worker` process); the job reschedules itself thereafter.
