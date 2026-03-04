# Middlewares, Logging, Testing, and Utilities

Read this file when writing middlewares, configuring logging, writing tests, or using
errutil/worker/mailer packages.

## Middlewares

Three layers of func nesting: (1) config → (2) chi handler wrapper → (3) handler.

```go
func RequirePermission(cfg *config.Source, perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !checkPermission(r.Context(), perm) {
				render.Error(w, r, 403, httperrors.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

For 3rd-party middleware, only the 2nd layer is needed — config layer wraps it.

### Default middlewares (via `AddDefaults()`)

- `middlewares.Configure` — injects `*config.Source` into request context
- `middlewares.LogRequests` — request/response metrics via httpsnoop
- `middlewares.CORSAllowAll()` — wraps `github.com/rs/cors`
- `middlewares.AddDataContext` — injects `*sqlx.DB` into request context

### Other middlewares

- `middlewares.CheckMigrations` — block routes until migrations applied (`MUST_MIGRATE=1`)
- `middlewares.DebugRequest` — dump req/resp bodies (`DEBUG_REQUEST=1` / `DEBUG_RESPONSE=1`)
- `middlewares.Sentry` — error reporting via `API_SENTRY_DSN`

### Selective application

Use chi `Route` and `Group`:

```go
r.Route("/todos", func(r chi.Router) {
	r.Get("/", c.List)
	r.Group(func(r chi.Router) {
		r.Use(auth.RequirePermission(cfg, "manage_own_todos"))
		r.Post("/", c.Create)
	})
})
```

## Logging (fxlog)

Structured logging via zerolog. Three levels:

```go
fxlog.Log("something happened")
fxlog.Error(err)           // also: fxlog.Errorf("failed: %w", err)
fxlog.Fatal(err)           // also: fxlog.Fatalf("fatal: %w", err)
```

Switch sink: `LOG_SINK=slog` to redirect to `log/slog`. Or set custom:

```go
fxlog.SetSink(fxlog.NewSlogSink(slog.Default()))
fxlog.SetSink(fxlog.NewZerologSink(customLogger))
```

## Testing (fxtest)

- `fxtest.Configure()` — returns `*config.Source` for tests
- `fxtest.ConnectTestDatabase(t)` — returns `context.Context` with config + `*sqlx.DB`;
  database auto-dropped on test completion unless `FXTEST_CLEANUP=no`

```go
func TestSomething(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	err := data.Exec(ctx, "INSERT INTO todos (title) VALUES ($1)", "test")
	require.NoError(t, err)
}
```

## Error Utilities (errutil)

- `errutil.Wrap(name, &err)` — defer; prefix error with name
- `errutil.WithCode(err, code)` — attach string error code (API responses)
- `errutil.WithData(err, data)` — attach arbitrary context data
- `errutil.NewCoded(code, msg, data)` — new error with code + message + data
- `errutil.Decorate(err)` — wrap for JSON serialization
- `errutil.Aggregate[T](slice, func)` — parallel execution, collect errors
- `errutil.AggregateWithTags[T](slice, func)` — like Aggregate with labels

## Background Workers

PostgreSQL-backed job system.

```go
w := worker.New(cfg, &SendEmailJob{}, &CleanupJob{})
w.Start() // blocks, polls for jobs
w.Stop()  // graceful shutdown
```

Job interface:

```go
type Interface interface {
	Name() string
	Run(ctx context.Context) error
}
```

Scheduling:

```go
worker.ScheduleNow(ctx, &SendEmailJob{To: "user@example.com"})
worker.ScheduleIn(ctx, &CleanupJob{}, 30*time.Minute)
worker.ScheduleAt(ctx, &ReportJob{}, tomorrow)
worker.ScheduleNowIfNotExists(ctx, &DailyDigestJob{})
```

Config: `WORKER_POLL` — polling interval (default: `1m`).

## Mailer

Transactional email via Postmark.

```go
err := mailer.Send(cfg, &mailer.Mail{
	From:     "noreply@example.com",
	To:       []string{"user@example.com"},
	Subject:  "Welcome!",
	HTMLBody: "<h1>Hello</h1>",
	TextBody: "Hello",
})
```

Config: `POSTMARK_TOKEN` — Postmark server API token.
