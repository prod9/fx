# Pub/Sub — Postgres LISTEN/NOTIFY

**Status:** accepted (not yet implemented)

The `pubsub` package is a typed pub/sub built on Postgres `LISTEN`/`NOTIFY`. FX apps are
already locked into Postgres, so pub/sub comes from the database already running — no new
broker, no NATS, no Redis. It sits next to `worker` and `cache` as a top-level package.

A channel is declared once at package scope, generic over its payload type — the
`config.*Var` pattern — so publish and subscribe are type-checked end to end. `Subscribe`
hands the app a native Go channel to `for range` over, backed by one dedicated `LISTEN`
connection; the package hides the connection, the goroutine, and reconnection.

It is a **latency optimization over polling, not a delivery guarantee.** Apps stay correct
without it and merely react faster with it; the package adds no reliability machinery on
top of Postgres. Anything needing guaranteed delivery uses `worker`, not `pubsub`.

The model is deliberately flat: **one `Subscribe` = one connection = one channel**, with
Postgres doing cross-subscriber fan-out natively (§4). Two application modes drive the
shape, both first-class:

- **Stand-alone listener** — a consumer process that ranges over notifications and acts,
  the pull-based analog of today's polling `worker`.
- **Websockets** — an HTTP handler that, per connected client, subscribes with the request
  context and forwards notifications out over the socket.

## The governing constraint — signaling, not a queue

`NOTIFY` is ephemeral, at-most-once:

- Dropped if no session is currently listening — never persisted.
- Dropped during any reconnect gap (a connection blip loses messages).
- Payload must be **under 8000 bytes** (`< 8000`; the limit is exclusive).
- Delivered only on transaction commit; an aborted tx sends nothing.
- Ordered per channel — but *at Postgres*: send order within a tx, commit order across
  tx. This is a property of the server queue, not a guarantee the consumer sees, since
  `pubsub` drops to slow consumers and reconnects across blips (§4, §5).
- De-duplicated within a single transaction: identical `(channel, payload)` notifications
  raised more than once in one tx are collapsed to one delivery.
- Scoped to one **database**: `NOTIFY` never crosses databases, so per-environment DBs
  (dev / staging / prod on separate databases) are fully isolated with no channel-name
  coordination needed.

It is **not** a durable job queue — that is `worker`. Design apps to use it as a
*notification*, never as payload delivery (see Recommended usage).

## Recommended usage — a notification, not a delivery channel

`pubsub` is rooted in Postgres's own mechanism and adds as little fail-safe on top as
possible. Delivery is best-effort by contract; **correctness is the app's job**, designed
around two rules.

**1. Publish out-of-band from the business write.** Commit the state change in its own
transaction, then publish *after* it — not inside it. A full notify queue makes `NOTIFY`
fail at commit (§5); coupling it into the business tx would let that roll back real data.
Only publish inside the business tx when missing the notification is genuinely worse than
failing the write — rare.

```go
if err := data.Run(ctx, saveOrder); err != nil {   // durable change commits first
    return err
}
if err := pubsub.Publish(ctx, OrdersChanged, OrderEvent{OrderID: id}); err != nil {
    // out-of-band: this failure never touches the committed order — log and move on
    fxlog.Warn(ctx, "orders_changed notify failed", "err", err)
}
```

**2. Subscribe as a wake-up, then re-derive from the table.** Treat a notification as
"something changed, look now" and keep the source of truth in the table. The consumer
advances a **cursor over the table's natural ordering** — re-scanning `WHERE status = …
AND id > $lastSeen` and remembering the highest id it processed — backstopped by a
periodic scan on a slow timer. The notification only collapses latency from "next periodic
scan" to "now"; a lost notification costs latency, never correctness, because the cursor
still sweeps every unprocessed row on the next tick.

```go
ch, cancel, err := pubsub.Subscribe(ctx, OrdersChanged)
if err != nil {
    return err
}
defer cancel()

lastSeen := int64(0)
scan := func() { lastSeen = rescanShipping(ctx, lastSeen) }   // cursor over natural order
scan()                                                          // catch up on startup
ticker := time.NewTicker(30 * time.Second)                     // periodic backstop
defer ticker.Stop()
for {
    select {
    case <-ch:        scan()          // "something changed, look now"
    case <-ticker.C:  scan()          // backstop — covers any missed notification
    case <-ctx.Done(): return ctx.Err()
    }
}
```

Do **not** narrow the scan to only the id carried in the payload — that reintroduces the
delivery dependence the cursor exists to remove. A monotonic sequence column lets the
cursor additionally *skip* already-seen rows as an optimization, but the full window sweep
stays the correctness floor. Because the consumer re-derives from the table, payloads stay
tiny — an id, or nothing — which also keeps clear of the size limit.

## Why the existing surfaces don't fit as-is

- `data` exposes only the pooled `*sqlx.DB` and an always-in-a-transaction `Scope`
  (`data/scope.go`). A listener must hold one connection open and block on it, outside any
  transaction — the pooled/tx model cannot express that. The connection still comes from
  the pool `data.Connect` builds (§4); no new connection API is needed.
- `worker` polls on a timer and holds no connection between polls (`worker/worker.go`) — a
  useful CLI template, the wrong connection model.

## Design

### 1. Channels — declared once, typed, immutable descriptors

Declared at package scope next to their consumer, generic over the payload type.
`Channel[T]` is an immutable descriptor (a name and a phantom `T`) holding no live state;
every operation routes through the `pubsub` package, mirroring `config.Get(src, Var)`:

```go
type OrderEvent struct { OrderID int64; Status string }

var OrdersChanged = pubsub.NewChannel[OrderEvent]("orders_changed")
```

**Signal-only channels** — where the fact of the event is the whole message and there is
no payload — use the empty struct as the convention; there is no separate `NewSignal`
alias:

```go
var CacheFlushed = pubsub.NewChannel[struct{}]("cache_flushed")
```

`NewChannel[T]` records a type-erased descriptor (name + type name) in a package registry
that powers `pubsub channels` introspection. **It panics on a duplicate registration** —
the same channel name declared twice, or the same name with a different `T` — because a
name collision is a program-construction error that must surface at startup, not a
runtime condition to handle. Payloads ride the wire as JSON of `T`; that boundary is the
one sanctioned `any`-shaped seam.

**Channel names are Postgres identifiers.** `LISTEN` / `NOTIFY` cannot parameterize the
channel name, so `NewChannel` validates the name as a plain identifier (letters, digits,
underscore) up front — a default-deny check that closes the only injection seam in the
package. Note Postgres truncates identifiers at 63 bytes (`NAMEDATALEN`): two names
sharing a 63-byte prefix silently collide onto one channel, so keep names short and
distinct.

### 2. Publish — rides the transaction

```go
func Publish[T any](ctx context.Context, ch Channel[T], payload T) error
```

`ctx` leads, per Go convention. Publish runs through the existing `data.Exec` path on the
pooled connection, so it rides whatever tx context it is called in: inside `data.Run(...)`
it joins the parent tx and fires on that commit; called on its own it runs a
`BEGIN` / `pg_notify` / `COMMIT` of its own (three round-trips) and fires on that commit.
Prefer the out-of-band form, after the business commit (Recommended usage rule 1) — since
coupling a `NOTIFY` into the business tx lets a full notify queue roll it back. The
marshaled payload must be under 8000 bytes, else `Publish` returns an error. No schema, no
migration.

### 3. Subscribe — a bare typed channel, its cancel, and a connect error

```go
func Subscribe[T any](ctx context.Context, ch Channel[T]) (<-chan T, context.CancelFunc, error)
```

`Subscribe` opens its own dedicated connection, issues `LISTEN`, and runs one
`WaitForNotification` loop that decodes each payload into `T` and feeds the returned
channel. **The initial connect and `LISTEN` are synchronous: if either fails — including a
connection-cap block or Postgres "too many clients" (§4) — `Subscribe` returns the error
rather than a live channel**, so the failure surfaces loud at the callsite instead of a
silent never-delivering stream.

On success it returns the channel **and** its cancel, so the caller never builds a
cancellable context of their own. A subscription owns its connection alone, so its whole
lifetime is one context: cancelling — via the passed `ctx` (a websocket request dying on
disconnect) or the returned `cancel` — stops the loop, closes the channel, and releases
the connection.

The channel carries bare `T` — no wrapper, no gap signal, no per-message error. That is
the point: `pubsub` is a latency optimization, not a delivery guarantee, and the app is
built so missed notifications never cost correctness (Recommended usage). A payload that
fails to decode into `T` is logged and skipped; the consumer is unaffected because it
re-derives from the table and never had to trust the payload.

`cancel()` is safe mid-range and idempotent — the loop goroutine is the sole sender and
closes the channel only after `WaitForNotification` unblocks on the cancelled context, so
a consumer-initiated cancel never races a send. `break` after `cancel()` to stop at once;
otherwise the range ends when the close propagates. A CLI wires the same cancel to CTRL-C
with `ctrlc.Do(cancel)` — no `context.WithCancel` of its own.

**A subscription must be cancelled.** Abandoning the range without calling `cancel` (and
without the `ctx` being cancelled) leaks the loop goroutine *and* its held Postgres
connection for the life of the process — the connection is never returned to the pool.
`defer cancel()` right after a successful `Subscribe` is the standing pattern.

**Websockets:** after `Hijack()`, the request's `r.Context()` is no longer cancelled on
client disconnect — the server stops managing the connection. The read loop is the
disconnect detector, so the handler must call the returned `cancel` when the socket read
fails, rather than relying on `r.Context()` to tear the subscription down.

### 4. One connection per subscription — Postgres does the fan-out

Each `Subscribe` pulls one dedicated connection from the pool `data.Connect` already
produced (`db.Conn(ctx)`) and holds it for the subscription's life, listening on it with
`WaitForNotification` — no `data.Dial` / new connection API. Reusing the one pool keeps
connection accounting honest: subscriptions draw against the same `DATABASE_MAX_OPEN`
budget as every other query, rather than hiding
in a shadow pool that makes the configured limit lie. The DB comes from the `data`
context, so there is no process-wide service to register: websocket handlers already have
it on the request context, and a stand-alone command builds a data context the way other
offline commands do.

Cross-subscriber fan-out is Postgres's job: every connection `LISTEN`ing on a channel
receives every `NOTIFY`, so N subscribers are N connections and the database delivers to
all. The package builds no in-memory fan-out this phase.

On a `WaitForNotification` error the loop distinguishes a cancelled context (a clean
shutdown — stop and release) from a dropped connection (reconnect with tight backoff and
re-`LISTEN`). Reconnect is silent: the blip's gap is invisible to the consumer, which the
app's re-scan (Recommended usage) already covers. Reconnect is liveness, not a reliability
upgrade — it resumes listening, it does not recover the messages missed during the outage.
Before a connection is released back to the pool the subscription issues `UNLISTEN *`, so
a reused connection never carries a stale registration into its next borrower.

**Cost — connection count equals active subscriptions.** One subscription is one Postgres
backend (~5–10 MB of server memory each), so the practical ceiling is **low hundreds**,
bounded by `max_connections` and `DATABASE_MAX_OPEN`. A handful of stand-alone consumers
is nothing; thousands of websocket clients is thousands of connections and does not fit
this model. High websocket concurrency waits on the deferred in-process multiplexing
(Future work) or moves to a real broker (NATS) — not raw Postgres `LISTEN`.

Because subscriptions share the pool, an app that runs them **must** budget for them.
`DATABASE_MAX_OPEN` defaults to `64` (see Configuration) precisely so a subscription
cannot exhaust an unbounded pool; size it against expected subscription count plus normal
query load.

### 5. Backpressure — Postgres's semantics, not our subsystem

No custom buffering policy: no `PUBSUB_*` buffer knob, no drop metric, no gap signal, no
coalescing or keep-newest contract. The loop always drains the connection and sends to the
consumer non-blocking; a consumer that can't keep up silently misses messages — the same
at-most-once contract Postgres has when no one is listening, and covered by the app's
re-scan.

Blocking the loop to force true Postgres-side backpressure is rejected as a process-wide
footgun. If a subscription stops reading its connection, notifications back up in
Postgres's single shared server-side async queue; once that global queue fills, **every
`NOTIFY` transaction fails at commit, across all channels and sessions**
(`pg_notification_queue_usage()` warns as it approaches). One stalled consumer must never
fail unrelated publishers, so the loop drains always and drops to the slow consumer.

What this costs the caller is understanding the Postgres channel contract by convention —
keep transactions short, keep consumers fast, treat delivery as best-effort — which this
spec documents rather than the package hides.

## Configuration

`pubsub` adds no config vars of its own. It depends on one `data` setting:

* `DATABASE_MAX_OPEN` — maximum open pooled connections, default `64`. Each active
  subscription holds one connection from this pool for its lifetime, so this value is the
  hard ceiling on concurrent subscriptions plus in-flight queries. The default replaces
  the previous unlimited (`-1`) setting with an explicit, honest budget; raising it is a
  deliberate capacity decision bounded by Postgres `max_connections`.

> This default change lives in `data` (`data/data.go`), blast radius = every FX app, and
> ships with the `pubsub` implementation.

## Operational — debugging is first-class

A `pubsub` cobra command group mirroring `store`:

| Command                        | Purpose                                                         |
|--------------------------------|----------------------------------------------------------------|
| `pubsub notify <channel> <payload>` | Publish from the CLI.                                      |
| `pubsub listen <channel>...`   | Subscribe and print to stdout; also the reference stand-alone consumer. |
| `pubsub channels`              | List declared channels from the `NewChannel` registry.         |

The CLI operates on channel *names* as strings — the untyped floor beneath the typed
`Channel[T]` layer — since a command line has no compile-time `T`. The typed API is the
sanctioned path for application code; the string path exists for this debug surface.

`pubsub` ships no metrics or counters subsystem: delivery is best-effort and silent by
design, and the CLI plus Postgres's own `pg_notification_queue_usage()` are the
introspection surface. Adding counters is explicitly out of scope — there is no reliable
number to report under an at-most-once contract, and the app's re-scan is where
correctness is observed.

## Testing

The typed marshal/registry layer and channel-name validation are unit-testable without a
database. Publish/subscribe round-trips test against a real Postgres (`fxtest`
`ConnectTestDatabase`), asserting best-effort delivery under a live listener and the
documented drop under a stalled consumer — not exactly-once, which the contract does not
promise.

**Payload evolution across deploys:** because payloads are JSON of `T` and old and new
binaries can run concurrently during a rollout, treat `T` like any wire schema — add
fields, never repurpose or remove them within a compatibility window, and let unknown
fields decode away. A consumer that re-derives from the table (Recommended usage) is
resilient to a payload it can't fully decode anyway.

## Future work (out of scope this phase)

- **Connection multiplexing / in-process fan-out** — a shared per-process connection that
  `LISTEN`s once per channel and fans out in memory to many subscribers, so
  websocket-scale consumer counts stop mapping 1:1 to Postgres connections. Deferred until
  the usage pattern is real. Note `PgBouncer` is not the escape valve: transaction /
  statement pooling disables `LISTEN`/`NOTIFY` entirely, and session pooling pins one
  server connection per client — zero multiplexing either way.
- **Dynamic channel names** — per-tenant `orders:42` / per-user `user:7` don't fit a
  static declaration. Options, undecided: a `pubsub.Bind(ch, suffix)` returning a same-`T`
  channel on `name:suffix`, or a lower-level string subscribe beneath the typed layer.
  Under the flat model these are just more subscriptions, so this rides on the
  multiplexing work above.
- **`Fanout` helper** — `pubsub.Fanout(ch)` splitting one subscribed stream to several
  in-process consumers. Consumer-side sugar on `Subscribe`; shape depends on a
  pattern-of-use we haven't seen. Deferred.

## Philosophy check

| Principle                    | How it holds                                                                          |
|------------------------------|---------------------------------------------------------------------------------------|
| Modular by composition       | ships as a package of typed channel descriptors, not a bus                            |
| Thin wrapper over primitives | wraps `pg_notify` + `WaitForNotification`; pg semantics surfaced, not hidden           |
| Decentralized declaration    | channels declared at package scope like config vars; self-register for introspection  |
| Context as carry-bag         | publish and subscribe both read the ambient `data` context                            |
| Everything optional          | no wiring for callers; no process-wide service to start                               |
| Operational first-class      | `pubsub notify` / `listen` / `channels` subcommands                                   |
| Punt distributed problems    | at-most-once documented; durability points at `worker`; fan-out leans on Postgres      |
| Convention one layer deep    | one subscribe surface (a typed stream); no handler/callback variant to choose          |
