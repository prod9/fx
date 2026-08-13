# Pub/Sub driver: Postgres `LISTEN`/`NOTIFY`

**Status:** accepted

The default driver for [`pubsub`](pubsub.md), active when `PUBSUB_URL` is unset. FX apps
are already locked into Postgres, so pub/sub comes from the database already running — no
new broker. The driver takes no URL of its own: it rides the `*sqlx.DB` already on the
`data` context, so there is no process-wide service to register — websocket handlers
already have it on the request context, and a stand-alone command builds a data context
the way other offline commands do. (`PUBSUB_URL=postgres://…` pointing at a *different*
database than `DATABASE_URL` is out of scope; the driver exists to reuse the database
already running.)

## Where the portable contract's limits come from

The package-wide limits in [`pubsub.md`](pubsub.md) originate here and are enforced on
every driver so channel declarations stay portable:

- **Payload under 8000 bytes** — the Postgres `NOTIFY` payload ceiling (exclusive).
- **Channel names are plain identifiers, at most 63 bytes** — `LISTEN` cannot
  parameterize its channel name, so the name is interpolated into SQL; the default-deny
  identifier whitelist closes the package's one injection seam. The 63-byte cap is
  `NAMEDATALEN`: Postgres truncates longer identifiers, so two names sharing a 63-byte
  prefix would silently collide onto one channel.

## Publish — rides the transaction

The send runs through `data.Exec` (`SELECT pg_notify($1, $2)`), so it rides whatever tx
context it is called in: inside `data.Run(...)` it joins the parent tx and fires on that
commit; called on its own it runs a `BEGIN` / `pg_notify` / `COMMIT` of its own (three
round-trips) and fires on that commit. An aborted tx sends nothing.

This is why the out-of-band rule (publish *after* the business commit) matters doubly
here: a full server-side notify queue makes `NOTIFY` fail at commit (see Backpressure),
and a publish coupled into the business tx would let that failure roll back real data.

The payload floor is `string`, not `[]byte`, because a `NOTIFY` payload is a `text`
column; a caller with binary data encodes it (base64, hex) into text itself.

Driver-specific extras — real on this driver, **not** part of the portable contract, and
not to be depended on by apps:

- Ordered per channel *at Postgres* (send order within a tx, commit order across tx) —
  a property of the server queue, not something the consumer sees, since the driver drops
  to slow consumers and reconnects across blips.
- De-duplicated within a single transaction: identical `(channel, payload)` notifications
  raised more than once in one tx collapse to one delivery.
- Scoped to one **database**: `NOTIFY` never crosses databases, so per-environment DBs
  (dev / staging / prod on separate databases) are fully isolated with no channel-name
  coordination needed.

## One connection per subscription — Postgres does the fan-out

Each `Subscribe` pulls one dedicated connection from the pool `data.Connect` already
produced (`db.Conn(ctx)`) and holds it for the subscription's life, listening with
`WaitForNotification` — no `data.Dial` / new connection API. Reusing the one pool keeps
connection accounting honest: subscriptions draw against the same `DATABASE_MAX_OPEN`
budget as every other query, rather than hiding in a shadow pool that makes the
configured limit lie.

Cross-subscriber fan-out is Postgres's job: every connection `LISTEN`ing on a channel
receives every `NOTIFY`, so N subscribers are N connections and the database delivers to
all. The driver builds no in-memory fan-out.

On a `WaitForNotification` error the loop distinguishes a cancelled context (a clean
shutdown — stop and release) from a dropped connection (reconnect with tight backoff and
re-`LISTEN`). Reconnect is silent: the blip's gap is invisible to the consumer, which the
app's re-scan already covers. Reconnect is liveness, not a reliability upgrade — it
resumes listening, it does not recover the messages missed during the outage. Before a
connection is released back to the pool the subscription issues `UNLISTEN *`, so a reused
connection never carries a stale registration into its next borrower.

**Cost — connection count equals active subscriptions.** One subscription is one Postgres
backend (~5–10 MB of server memory each), so the practical ceiling is **low hundreds**,
bounded by `max_connections` and `DATABASE_MAX_OPEN`. A handful of stand-alone consumers
is nothing; thousands of websocket clients do not fit this driver — that is what the
[Redis driver](pubsub-redis.md) is for.

Because subscriptions share the pool, an app that runs them **must** budget for them.
`DATABASE_MAX_OPEN` defaults to `64` (see `data`) precisely so a subscription cannot
exhaust an unbounded pool; size it against expected subscription count plus normal query
load.

## Backpressure — why the drop rule is non-negotiable here

The package-wide rule (drain always, drop non-blocking to a slow consumer) exists because
of this driver's failure mode. If a subscription stops reading its connection,
notifications back up in Postgres's single shared server-side async queue; once that
global queue fills, **every `NOTIFY` transaction fails at commit, across all channels and
sessions** (`pg_notification_queue_usage()` warns as it approaches). One stalled consumer
must never fail unrelated publishers, so the loop drains always and drops to the slow
consumer.

Introspection: `pg_notification_queue_usage()`, plus the `pubsub listen` / `notify` CLI.

## Future work

**Connection multiplexing / in-process fan-out** — a shared per-process connection that
`LISTEN`s once per channel and fans out in memory to many subscribers, so websocket-scale
consumer counts stop mapping 1:1 to Postgres connections. Partly superseded by the Redis
driver (which multiplexes at the broker); deferred until the usage pattern is real. Note
`PgBouncer` is not the escape valve: transaction / statement pooling disables
`LISTEN`/`NOTIFY` entirely, and session pooling pins one server connection per client —
zero multiplexing either way.
