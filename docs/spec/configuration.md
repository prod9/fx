# Configuration

**Status:** accepted

The `config` package allows defining configuration via environment variables in a way
that you don't end up with one giant `struct` and each module can maintain its own list
of configurable things.

Most modules in `fx` accept either a `context.Context` (which is expected to carry a
`*config.Source` inside it) or a `*config.Source` directly.

Start by creating a configuration variable:

```go
var SessionTimeoutConfig = config.DurationDef("SESSION_TIMEOUT", 3*time.Minute)
```

Available definitions include:

* `config.Str` — Plain string
* `config.Int` — Integer
* `config.Int64` — 64-bit integer
* `config.URL` — URL (wraps `url.Parse`)
* `config.Bool` — Boolean (wraps `strconv.ParseBool`)
* `config.Duration` — Duration (wraps `time.ParseDuration`)

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

You can use `config.Configure()` to just straight up get a new source to work with when
no context is available.

You can also `config.Set` or `config.SetDefault` to override values where needed, such
as in tests or when defaults in some modules are not what you want.

```go
// permanent overrides
config.Set(data.DatabaseMaxIdleConfig, 2)

// just the defaults, if not specified
config.SetDefault(data.DatabaseURLConfig, "postgres:///mydb")
```

Configurations are read from the following list, in order, with higher ones overriding
lower ones:

* `config.Set` — Values overridden at runtime.
* Actual Environment variables
* `.env.local` and `.env` files — searched from the current directory upward, stopping
  at the nearest `.git` directory. At each level, `.env.local` overrides `.env`, and
  closer files override those further up. Useful for monorepo setups with shared envs.
* `config.SetDefault` — Default values overridden at runtime.
* `config.*Def` — Default values set on definition.
* Go defaults (e.g. `0` for int, `""` for string, etc.)

## Configuration problems never block startup

A configuration problem must not prevent the app from starting. Config is read before
anything else, so a hard failure there would also take down the tools needed to diagnose
it (`print-config`, `data psql`, health routes). Instead:

* An unset or blank value yields the default.
* A value that does not parse logs `<NAME>: <error>` and yields the default.
* A failed provider read (for example the `settings` provider's database is down) logs
  `config: <error>` and yields the default.

A consumer that cannot work without a valid value fails where it uses it: a wrong
`DATABASE_URL` still fails when the app connects. `config.GetOK` reports `false` in every
case above, for callers that must tell "set" from "defaulted".

## Aliases

A variable can be read from older names, so a rename does not break deployments that
still set the old one:

```go
var SentryDSNConfig = config.Str("SENTRY_DSN").WithAlias("API_SENTRY_DSN")
```

`Name()` is always the canonical name; `Aliases()` lists the older names. Reads try the
canonical name first, then each alias in order, taking the first non-blank value, so the
canonical name wins when both are set. The first read through an alias logs once that the
variable should be renamed. `config.Set` writes the canonical name only.
`print-config <name>` accepts the canonical name or any alias.

fx's own variables dropped their `API_` prefix this way: `API_SENTRY_DSN` is now
`SENTRY_DSN` and `API_CORS_ORIGINS` is now `CORS_ORIGINS`; the old names still work as
aliases.

## Conventions for App-Level Config

A few config values are needed by almost every API but are intentionally not built into
fx because their resolution shape varies too much across apps (multi-tenant, reverse
proxies with `X-Forwarded-*`, per-environment domains). For these, fx publishes a
recommended env var name so that apps using fx stay consistent with each other without
fx itself owning the value.

* `API_PREFIX` — the public base URL of the API, used for OAuth redirect URIs,
  absolute links in email, webhook callback URLs, presigned URL hostnames, etc. The
  name is deliberate: "prefix" tells you how to use it — concatenate, as in
  `API_PREFIX + "/auth/callback"`. fx does not read this variable; declare and read it
  in your app with the normal `config` package idioms.
