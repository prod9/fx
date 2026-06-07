# Middlewares

**Status:** accepted

Middlewares in `fx` attempt two things: conform to Go's standard `http.Handler`
signature and be easy to use with `github.com/go-chi/chi`.

They are a bit complicated, but not hard to understand. There are 3 layers of `func`
nesting:

1. First layer takes in a configuration to configure the middleware.
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

* The first layer is expected to be called multiple times throughout your application
  to create different instances of the middleware. You will need to be careful with
  handling any global state, if needed.

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

* `middlewares.Configure` — Injects `*config.Source` into the request context.
* `middlewares.LogRequests` — Captures metrics with `github.com/felixge/httpsnoop` and
  logs requests/responses.
* `middlewares.CORSAllowAll()` — Wraps `github.com/rs/cors` to allow all CORS requests.
* `middlewares.AddDataContext` — Adds a `*sqlx.DB` to the request context.

Some other non-default middlewares include:

* `middlewares.CheckMigrations` — If `MUST_MIGRATE=1` is set, all routes will return
  errors until all known migrations are applied.
* `middlewares.DebugRequest` — Prints out incoming request body if `DEBUG_REQUEST=1` is
  set, or prints out the outgoing response body if `DEBUG_RESPONSE=1` is set.
* `middlewares.Sentry` — Installs sentry error reporting handler with DSN set in
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
