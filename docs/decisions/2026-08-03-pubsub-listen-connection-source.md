# pubsub listen connection — pooled via `data.Connect`, not a separate pgx connection

- **Date:** 2026-08-03
- **PR:** manual
- **Status:** accepted

## Decision

A `pubsub` subscription's long-lived `LISTEN` connection is a dedicated connection drawn
from the one pool `data.Connect` already builds (`db.Conn(ctx)`), held for the
subscription's life and returned to that pool on cancel. It is **not** a separate
`pgx.Connect`, and **not** a second `pgxpool` sized on its own. To accommodate held
subscription connections, `DATABASE_MAX_OPEN` changes its built-in default from `-1`
(unlimited) to `64`. That default bump is the only connection-model change pubsub ships.

## Rationale

The load-bearing property is **one honest connection budget**. Every FX app is already
locked into a single `*sqlx.DB` pool whose ceiling is `DATABASE_MAX_OPEN`, and that number
is what operators size against Postgres `max_connections`. A subscription holds one backend
for its whole life, so subscriptions are connections too — and the only way that stays
honest is if they draw against the *same* budget as every query. Any connection pubsub
opens outside the pool is invisible to `DATABASE_MAX_OPEN`: the configured limit then lies,
and an app can silently walk past Postgres `max_connections` and fall over at the worst
possible moment. Reusing the pool makes the cost visible at config time; the `-1 → 64`
default replaces a dishonest infinite budget with a finite, explicit one.

Reaching the pgx handle is unavoidable and is **not** what this decision forbids.
`WaitForNotification` is a `*pgx.Conn` method with no database/sql equivalent, and pgx is
already our driver. The subscription holds a pooled `*sql.Conn` (`db.Conn(ctx)`) and
reaches its underlying `*pgx.Conn` through the `pgx/v5/stdlib` `.Raw()` escape hatch —
pgx's own public, supported bridge for `LISTEN`/`NOTIFY` and `COPY`. The connection stays
owned and counted by the pool; only the always-in-a-transaction `data.Scope` model is
bypassed for the bare listen loop, which is why the spec's "why the existing surfaces don't
fit" section calls this out as a deliberate exception rather than a new connection API.

## Why not the obvious alternatives

- **A separate `pgx.Connect` per subscription.** Simpler at the callsite, but it opens a
  connection the pool can't see. `DATABASE_MAX_OPEN` no longer bounds real backend usage,
  so the one number operators trust becomes a lie and `max_connections` exhaustion goes
  silent until it isn't. Simplicity of one function is not worth an operational footgun.

- **A dedicated `pgxpool` for pubsub, sized on its own knob.** Two pools mean two numbers
  to size, and it is their *sum* — not either alone — that hits `max_connections`. A second
  pool doesn't remove the ceiling; it splits it and makes the real limit implicit. One pool
  is one budget an operator already understands, and it fits FX's "convention one layer
  deep / punt capacity to infra" stance.

- **NATS / Redis / a real broker.** Adds infrastructure to a stack already committed to
  Postgres, for a feature that is explicitly a best-effort latency optimization over
  polling, not a delivery guarantee. Guaranteed delivery already has a home: `worker`.

## Objections a fresh agent will raise, and the answers

| Objection | Answer |
|-----------|--------|
| `WaitForNotification` needs `*pgx.Conn`, so you're forced into a separate pgx connection. | No — hold a pooled `*sql.Conn` from `db.Conn(ctx)` and reach the pgx handle via `sqlConn.Raw` + `driverConn.(*stdlib.Conn).Conn()`. The connection is never opened outside the pool. |
| A held listen connection starves the pool for real queries. | That is the point of budgeting it. `-1 → 64` makes the held cost finite and visible; an unlimited pool would only hide exhaustion until `max_connections` fails. |
| A dedicated `pgxpool` isolates subscription load — cleaner. | Two pools, two numbers, and the sum is the real ceiling. Isolation here trades one honest budget for two implicit ones. |
| `pgx.Connect` is fewer lines than the `.Raw()` type assertion. | Fewer lines at one callsite, dishonest system-wide. The `.Raw()` bridge is pgx's sanctioned public hatch, encapsulated once inside `Subscribe`. |
| Reaching pgx internals via `.Raw()` is fragile. | `stdlib.Conn` and its `.Conn()` method are exported, documented API intended for exactly `LISTEN`/`COPY`; it is not reflection into unexported internals, and it adds no dependency pgx isn't already. |
| Holding a connection outside a transaction breaks `data`'s tx-always model. | Deliberately, and only for the listen loop. The connection still comes from `data.Connect`'s pool; only `data.Scope` tx-scoping is bypassed, documented as an exception. |
| `64` is arbitrary. | It is a finite starting budget with headroom under a default Postgres `max_connections` of 100. It is not a tuning target — it is an honest default replacing an infinite one; raising it is a deliberate capacity decision. |

## Ties into

- Spec: [`../spec/pubsub.md`](../spec/pubsub.md) §4 (one connection per subscription) and
  Configuration (`DATABASE_MAX_OPEN` default `64`).
- Code: the default lives in `data/data.go` (`DatabaseMaxOpenConfig`), blast radius = every
  FX app, and ships with the pubsub implementation.
