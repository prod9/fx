# Design Philosophy

**Status:** current

The throughline that guides what FX is and — equally important — what it refuses to be.
Treat this as the design intent test for any new feature, package, or abstraction. If a
proposed change conflicts with one of the principles below, the burden is on the change
to argue why this case is the exception, not on the principle to defend itself.

## What FX is

A minimalistic, modular Go API framework. Install the normal way — `go get
fx.prodigy9.co` — and import the packages you need. FX is now mature enough that most
projects don't need to fork it.

`git subtree` remains a supported escape hatch for cases where you want to hack on FX
itself from inside a downstream app (faster turnaround than upstream PR + bump). It is
no longer the recommended default. If you reach for subtree, the expectation is that
fixes flow back upstream via `git subtree push`.

## Principles

### 1. Modular by composition, not by config

Applications are built by composing app fragments — `app.Build().Mount(auth.App).Mount(todo.App)`
— Django-app-style bundles of controllers + middlewares + commands. No DI container, no
plugin registry, no service locator. Just a builder that walks fragments and wires them
into chi + cobra.

**Implication:** new cross-cutting capabilities should ship as a fragment with its own
controllers/middlewares/commands, not as a registry or hook system.

### 2. Thin wrappers over standard primitives, never replacements

- Controllers are an interface around a `chi.Router`.
- Middlewares are `func(http.Handler) http.Handler`.
- The DB layer is `sqlx` with a transaction-scope helper.
- Commands are `cobra` commands.

You can always escape down to the underlying library; nothing is hidden behind a custom
abstraction. The framework adds ergonomics, not opacity.

**Implication:** prefer extending the standard primitive over inventing a new one. If
you reach for a custom interface, ask whether the underlying library already has one.

### 3. Config as decentralized declarations

`var SessionTimeout = config.DurationDef(...)` lives next to the code that reads it —
no central `Config` struct that every package fights over. Layered resolution (`Set` →
env → `.env*` → `SetDefault` → `*Def` → zero) is documented and predictable.

**Implication:** new tunables declare their own `config.*Def` next to their consumer.
Don't aggregate config into shared structs.

### 4. Context as the carry-bag

Both `*config.Source` and `*sqlx.DB` ride in `context.Context`. Functions take a `ctx`
and pull what they need. This is what lets transaction scopes propagate transparently
via `scope.Context()` — nested calls don't know they're in a transaction.

**Implication:** new ambient values (request-scoped state, tenant info, tracing) ride
the context. Don't introduce parallel passing mechanisms.

### 5. Everything is optional

`data` is opt-in. `worker` is opt-in. `mailer`, `cache`, `secret`, `blobstore` are
independent. `AddDefaults()` is convenience, not a contract. The framework refuses to
demand you adopt all of it.

**Implication:** new packages must work standalone. Cross-package coupling needs strong
justification — and even then, prefer making it optional with a sensible default.

### 6. Operational concerns are first-class, not afterthoughts

`print-config`, `data migrate`, `data psql`, `CI=1` / `ALWAYS_YES=1` flags,
`EmbedMigrations` via `go:embed` for prod, `MUST_MIGRATE` gate, Sentry middleware,
`__panic` route. The CLI exists to make a deployed binary debuggable without shelling
around.

**Implication:** any new subsystem should answer "how do I debug this in prod?" before
it ships. Often that means a new `cmd` subcommand or a built-in introspection route.

### 7. Punt hard distributed-systems problems to infrastructure

HA is "use pgbouncer." There's no service-mesh layer, no retry framework for HTTP, no
pluggable transport. The framework draws a line at the process boundary and lets
infrastructure handle what infrastructure handles better.

**Implication:** resist the urge to absorb distributed-systems concerns into the
framework. Document the infra assumption instead.

### 8. Convention over options, but only one layer deep

Three transaction styles (`Run`, `NewScopeErr`, `NewScope`) — one recommended, two
escape hatches. Two log sinks built in (zerolog, slog) plus a `Sink` interface. Enough
flexibility to not fight you, not enough to become a configuration tarpit.

**Implication:** when adding flexibility, cap it at one recommended path plus a small
number of escape hatches. Reject pluggable-everything designs.

## The throughline

FX reads like a framework written by someone who has been burned by both "magic"
frameworks (Rails-style autoloading, Spring-style DI) and DIY-everything Go projects.
The compromise: take the 10–15 things every API needs (config, routing, DB, migrations,
jobs, logging, mail, caching, secrets), give each a thin idiomatic wrapper, compose
them through a builder, and ship as an importable Go module that doesn't fight the
ecosystem.

Minimalism here is not a goal. It is a discipline against accidental complexity.
