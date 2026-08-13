# Pub/Sub — typed channels over pluggable drivers

**Status:** accepted (driver layer: draft)

The `pubsub` package is a typed pub/sub with a pluggable backend. The default backend is
Postgres `LISTEN`/`NOTIFY` ([`pubsub-postgres.md`](pubsub-postgres.md)) — no new broker,
pub/sub comes from the database already running. A Redis driver
([`pubsub-redis.md`](pubsub-redis.md)) ships alongside it for deployments that already
run Redis. The driver seam is internal — the two built-ins are the only backends, and an
app committed to another transport (NATS, a real broker) uses that system's client
directly instead of importing `pubsub`. It sits next to `worker` and `cache` as a
top-level package.

A channel is declared once at package scope, generic over its payload type — the
`config.*Var` pattern — so publish and subscribe are type-checked end to end. `Subscribe`
hands the app a native Go channel to `for range` over; the package hides the connection,
the goroutine, and reconnection.

It is a **latency optimization over polling, not a delivery guarantee.** Apps stay correct
without it and merely react faster with it; the package adds no reliability machinery on
top of the backend. Anything needing guaranteed delivery uses `worker`, not `pubsub`.

The model is deliberately flat: **one `Subscribe` = one backend subscription = one
channel**, with the backend doing cross-subscriber fan-out natively. Two application modes
drive the shape, both first-class:

- **Stand-alone listener** — a consumer process that ranges over notifications and acts,
  the pull-based analog of today's polling `worker`.
- **Websockets** — an HTTP handler that, per connected client, subscribes with the request
  context and forwards notifications out over the socket.

## The governing constraint — signaling, not a queue

Every sanctioned driver is ephemeral, at-most-once. The portable contract — the floor an
app may rely on regardless of driver — is:

- Dropped if no subscriber is currently listening — never persisted.
- Dropped during any reconnect gap (a connection blip loses messages).
- Dropped to a consumer that can't keep up (§5).
- Payload must be **under 8000 bytes** (`< 8000`; the limit is exclusive).
- Channel names are plain identifiers of at most 63 bytes (§1).

The two limits originate in the Postgres driver (see
[`pubsub-postgres.md`](pubsub-postgres.md)) and are enforced uniformly on every driver so
channel declarations stay portable — a driver swap can never break a name or a payload
that worked before. Ordering, transaction-coupling, and de-duplication are
**driver-specific extras**, documented in each driver's file; apps must not depend on
them.

It is **not** a durable job queue — that is `worker`. Design apps to use it as a
*notification*, never as payload delivery (see Recommended usage).

## Recommended usage — a notification, not a delivery channel

`pubsub` adds as little fail-safe on top of its backend as possible. Delivery is
best-effort by contract; **correctness is the app's job**, designed around two rules.

**1. Publish out-of-band from the business write.** Commit the state change in its own
transaction, then publish *after* it — not inside it. Publishing is not transactional in
general (on most drivers it fires immediately), and on the Postgres driver a publish
coupled into the business tx can even roll back real data (see
[`pubsub-postgres.md`](pubsub-postgres.md)). Publishing after the commit is the one
ordering that is correct on every driver.

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

`NewChannel[T]` validates the name (below) and **panics on a malformed one** — a
program-construction error that must surface at startup, not a runtime condition to
handle. Channels are declared at package scope and named after the var that holds them
(the `config` convention), so an accidental name collision is a non-issue; the package
keeps no registry and does no duplicate-detection. Payloads ride the wire as JSON of `T`;
that boundary is the one sanctioned `any`-shaped seam.

**Channel names are plain identifiers** (letters, digits, underscore; at most 63 bytes)
on every driver — the rule's origin and rationale live in
[`pubsub-postgres.md`](pubsub-postgres.md). Enforcing it uniformly means a channel
declaration is portable and a driver swap can never introduce a collision or an injection
surface.

### 2. Publish

```go
func Publish[T any](ctx context.Context, ch Channel[T], payload T) error
func PublishRaw(ctx context.Context, name string, payload string) error
```

`Publish` marshals `payload` to JSON and delegates to `PublishRaw`, the untyped floor that
takes a channel name and a pre-marshaled payload — for the CLI and any caller holding a
name and a string rather than a typed channel. The floor is `string`, not `[]byte` (a
Postgres-rooted choice — see [`pubsub-postgres.md`](pubsub-postgres.md)); a caller with
binary data encodes it (base64, hex) into text itself. `ctx` leads, per Go convention.
`PublishRaw` validates the name and the size limit, then hands the send to the resolved
driver (§4). The marshaled payload must be under 8000 bytes, else `Publish` returns an
error. No schema, no migration.

Whether a publish rides an ambient transaction is driver-specific — the Postgres driver
joins the tx context it is called in; others fire immediately. Apps follow Recommended
usage rule 1 and never depend on either behavior.

### 3. Subscribe — a bare typed channel, its cancel, and a connect error

```go
func Subscribe[T any](ctx context.Context, ch Channel[T]) (<-chan T, context.CancelFunc, error)
func SubscribeRaw(ctx context.Context, name string) (<-chan string, context.CancelFunc, error)
```

`SubscribeRaw` is the untyped floor: it validates the name and hands the subscription to
the resolved driver (§4), which streams the payload text — for the CLI and name-based
callers. `Subscribe` wraps it, decoding each payload into `T` and feeding the returned
typed channel. **The initial connect and subscribe are synchronous: if either fails,
`Subscribe` returns the error rather than a live channel**, so the failure surfaces loud
at the callsite instead of a silent never-delivering stream.

On success it returns the channel **and** its cancel, so the caller never builds a
cancellable context of their own. A subscription owns its backend resources alone, so its
whole lifetime is one context: cancelling — via the passed `ctx` (a websocket request
dying on disconnect) or the returned `cancel` — stops the loop, closes the channel, and
releases the resources.

The channel carries bare `T` — no wrapper, no gap signal, no per-message error. That is
the point: `pubsub` is a latency optimization, not a delivery guarantee, and the app is
built so missed notifications never cost correctness (Recommended usage). A payload that
fails to decode into `T` is logged and skipped; the consumer is unaffected because it
re-derives from the table and never had to trust the payload.

`cancel()` is safe mid-range and idempotent — the driver's loop goroutine is the sole
sender and closes the channel only after its wait unblocks on the cancelled context, so a
consumer-initiated cancel never races a send. `break` after `cancel()` to stop at once;
otherwise the range ends when the close propagates. A CLI wires the same cancel to CTRL-C
with `ctrlc.Do(cancel)` — no `context.WithCancel` of its own.

**A subscription must be cancelled.** Abandoning the range without calling `cancel` (and
without the `ctx` being cancelled) leaks the loop goroutine *and* whatever backend
resource it holds (a Postgres connection; a Redis subscription) for the life of the
process. `defer cancel()` right after a successful `Subscribe` is the standing pattern.

**Websockets:** after `Hijack()`, the request's `r.Context()` is no longer cancelled on
client disconnect — the server stops managing the connection. The read loop is the
disconnect detector, so the handler must call the returned `cancel` when the socket read
fails, rather than relying on `r.Context()` to tear the subscription down.

### 4. Drivers — an internal seam, resolved from config

The backend seam sits exactly at the untyped floor and is **internal to the package** —
the public API is `Publish`/`Subscribe` (and their `Raw` floor) alone, opaque about what
runs underneath. Everything above the seam — `Channel[T]`, JSON codec, name and payload
validation — is driver-agnostic; everything below it is one small unexported interface
(`Publish(ctx, channel, payload string) error` and the matching `Subscribe`, at the raw
string level because Go interfaces cannot carry generic methods, and because everything
generic is codec work no driver should duplicate). Names and payloads arrive at the
driver already validated.

There are no extension points: no exported driver type, no context override, no scheme
registry. The sanctioned backends are the two built-ins, hard-coded; an app committed to
another transport (NATS, a real broker) should use that system's client directly rather
than flatten it to this at-most-once contract — it simply doesn't import `pubsub`. The
in-memory fake driving the unit tests is in-package, injected through an unexported test
seam.

**Resolution.** `PublishRaw` and `SubscribeRaw` resolve their driver per call from
`PUBSUB_URL` on the `ctx` config source (mirrors `STORAGE_URL`):

- `postgres` / `redis` scheme — the matching built-in; the rest of the URL is the
  driver's to interpret. Any other scheme is a resolution error, surfaced by the first
  publish or subscribe.
- Unset — the Postgres driver over the ambient `data` context. This is the
  zero-configuration default and exactly the pre-driver behavior, so existing apps need
  no change.

Drivers resolved from `PUBSUB_URL` are constructed lazily and cached per-process per-URL
(the `cache.Redis` / `blobstore.Client` pattern), so resolution on the hot path is a map
lookup, not a dial.

### 5. Backpressure — drop to the slow consumer, never block the loop

No custom buffering policy: no `PUBSUB_*` buffer knob, no drop metric, no gap signal, no
coalescing or keep-newest contract. Every driver drains its backend and sends to the
consumer non-blocking; a consumer that can't keep up silently misses messages — the same
at-most-once contract the backend has when no one is listening, and covered by the app's
re-scan.

Blocking the loop to force backend-side backpressure is rejected as a process-wide
footgun — on the Postgres driver a blocked loop can fail unrelated publishers outright
(see [`pubsub-postgres.md`](pubsub-postgres.md)). One stalled consumer must never fail
unrelated publishers, so every driver drains always and drops to the slow consumer, and
the consumer-visible contract doesn't shift under a driver swap.

What this costs the caller is understanding the at-most-once contract by convention —
keep consumers fast, treat delivery as best-effort — which this spec documents rather
than the package hides.

## Drivers

| Scheme     | Doc                                        | Summary                                                        |
|------------|--------------------------------------------|----------------------------------------------------------------|
| *(unset)*  | [`pubsub-postgres.md`](pubsub-postgres.md) | Default. `LISTEN`/`NOTIFY` over the ambient `data` context.    |
| `redis`    | [`pubsub-redis.md`](pubsub-redis.md)       | Redis PUB/SUB; websocket-scale subscription counts.            |

## Configuration

* `PUBSUB_URL` — driver selection, default empty (Postgres over the `data` context). The
  scheme picks the built-in driver; the remainder is driver-defined. Declared in
  `pubsub` next to its consumer, per the config philosophy.
* `DATABASE_MAX_OPEN` (`data`) — under the Postgres driver, the hard ceiling on concurrent
  subscriptions plus in-flight queries; see [`pubsub-postgres.md`](pubsub-postgres.md).

## Operational — debugging is first-class

A `pubsub` cobra command group mirroring `store`:

| Command                             | Purpose                                                                 |
|-------------------------------------|-------------------------------------------------------------------------|
| `pubsub notify <channel> <payload>` | Publish from the CLI.                                                   |
| `pubsub listen <channel>...`        | Subscribe and print to stdout; also the reference stand-alone consumer. |

The CLI operates on channel *names* as strings — the untyped floor beneath the typed
`Channel[T]` layer — since a command line has no compile-time `T`. It resolves its driver
the same way application code does (§4), so `PUBSUB_URL=redis://… go run . pubsub listen x`
debugs the Redis path with no extra flags. The typed API is the sanctioned path for
application code; the string path exists for this debug surface.

`pubsub` ships no metrics or counters subsystem: delivery is best-effort and silent by
design, and the CLI plus each backend's own introspection (listed in each driver's doc)
are the surface. Adding counters is explicitly out of scope — there is no reliable number
to report under an at-most-once contract, and the app's re-scan is where correctness is
observed.

## Testing

The typed marshal layer, channel-name validation, and driver resolution are unit-testable
without a backend — against an in-package in-memory fake driver on an unexported seam.
Postgres round-trips test against a real Postgres (`fxtest` `ConnectTestDatabase`);
Redis round-trips against a real Redis where available, skipped otherwise. Both assert
best-effort delivery under a live listener and the documented drop under a stalled
consumer — not exactly-once, which the contract does not promise.

**Payload evolution across deploys:** because payloads are JSON of `T` and old and new
binaries can run concurrently during a rollout, treat `T` like any wire schema — add
fields, never repurpose or remove them within a compatibility window, and let unknown
fields decode away. A consumer that re-derives from the table (Recommended usage) is
resilient to a payload it can't fully decode anyway.

## Future work (out of scope this phase)

- **Dynamic channel names** — per-tenant `orders:42` / per-user `user:7` don't fit a
  static declaration. Options, undecided: a `pubsub.Bind(ch, suffix)` returning a
  same-`T` channel on `name:suffix`, or a lower-level string subscribe beneath the typed
  layer. Under the flat model these are just more subscriptions.
- **`Fanout` helper** — `pubsub.Fanout(ch)` splitting one subscribed stream to several
  in-process consumers. Consumer-side sugar on `Subscribe`; shape depends on a
  pattern-of-use we haven't seen. Deferred.
- Driver-specific future work (e.g. Postgres connection multiplexing) lives in each
  driver's doc.

## Philosophy check

| Principle                    | How it holds                                                                          |
|------------------------------|----------------------------------------------------------------------------------------|
| Modular by composition       | ships as a package of typed channel descriptors, not a bus                            |
| Thin wrapper over primitives | wraps `pg_notify` / Redis PUB/SUB; backend semantics surfaced, not hidden             |
| Decentralized declaration    | channels declared at package scope like config vars; `PUBSUB_URL` declared in-package |
| Context as carry-bag         | config and the default's DB ride `ctx`; nothing new to wire at composition time       |
| Everything optional          | zero config = Postgres over the existing data context; no service to start            |
| Operational first-class      | `pubsub notify` / `listen` work against whichever driver is configured                |
| Punt distributed problems    | at-most-once documented; durability points at `worker`; fan-out leans on backend      |
| Convention one layer deep    | one subscribe surface; one config var; no extension points                            |
