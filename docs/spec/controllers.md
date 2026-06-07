# Controllers

**Status:** accepted

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
usually end with a call to the `render` package:

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

## Rendering

The `render` package has the following methods:

* `render.Text` — Renders plain text
* `render.JSON` — Renders JSON
* `render.Redirect` — Redirects (307) to a URL
* `render.FileTransfer` — Transfers a file (force a download)
* `render.Error` — Renders error JSON

## Built-in Controllers

The `httpserver/controllers` package includes several ready-to-use controllers:

* `Home{}` — Mounts `GET /` returning `{"time": "..."}` with the current server time.
  Useful for deployment testing and basic health checks.
* `Debug{}` — Mounts `GET /__panic` which triggers a test panic. Useful for verifying
  panic recovery middleware and error reporting (e.g. Sentry).
* `StaticJSON(path, obj)` — Creates a controller that serves a static JSON object at
  the given path via `GET`.
* `FromFunc(path, handlerFunc)` — Wraps an `http.HandlerFunc` as a controller mounted
  at the given path.
* `FromHandler(path, handler)` — Wraps an `http.Handler` as a controller mounted at the
  given path.
