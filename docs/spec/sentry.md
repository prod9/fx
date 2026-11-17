# Sentry

**Status:** implemented

Package `fxsentry` reports errors and panics to Sentry. It takes pieces of `sentry-go`
and shapes them to fx's patterns rather than following sentry-go's prescribed usage.

## Opt-in

Sentry is off unless something activates it and `SENTRY_DSN` (alias `API_SENTRY_DSN`)
holds a valid DSN. Two entry points activate it:

* `middlewares.Sentry`, part of `DefaultForAPI()`: for API apps, setting `SENTRY_DSN`
  is the opt-in.
* `worker.SentryFilter`, one of the worker's default filters.

`fxsentry.Activate(cfg)` builds the one process-wide client from the `config.Source` of
the first call; later calls reuse it (first wins). It reports whether Sentry is on. An
unset DSN is silent; an invalid one is logged. Neither stops the app (see
configuration.md).

## Scopes and capture

The context carries only nested scopes, never a client:

```go
ctx, scope := fxsentry.NewScope(ctx) // inherits the scope already on ctx
scope.SetTag("tenant", tenantID)

fxsentry.Capture(ctx, err)           // report an error with ctx's scope
defer fxsentry.CapturePanic(ctx)     // recover, report, flush, re-panic
```

All three do nothing while Sentry is off. The HTTP middleware builds a scope per request
(with the request attached); `SentryFilter` builds one per job run (tagged `job`).

Every event takes its release and `component` tag from `buildinfo.FromContext(ctx)`
(see buildinfo.md), so sentry-go's own release fallbacks (`SENTRY_RELEASE`, CI
variables) never apply. There is no `SENTRY_RELEASE` support.

`fxsentry` knows nothing about HTTP or workers; each chain's integration lives with that
chain. `fxlog` does not report to Sentry.

## Checking it in production

Set `DEBUG_SENTRY=1` and request `GET <prefix>/panic` on the `controllers.Meta`
controller: the panic reaches Sentry through the middleware.
