# Pub/Sub driver: Redis PUB/SUB

**Status:** accepted

The Redis driver for [`pubsub`](pubsub.md), active on `PUBSUB_URL=redis://…` (any URL
`goredis.ParseURL` accepts, including `rediss://`). Built on native Redis PUB/SUB, which
has the identical at-most-once shape — no persistence, dropped when no one listens,
dropped across reconnects — so the portable contract transfers unchanged. For the app
that already runs Redis, this moves pubsub traffic off the database's connection budget.

- One `go-redis` client per URL per process, acquired from `clients/redis` — the single
  place FX dials Redis, shared with `cache` (the same URL in `REDIS_URL` and `PUBSUB_URL`
  means one client, one pool). The client dials lazily on first command; each `Subscribe`
  opens one `*goredis.PubSub` on it. Redis connections are far cheaper than Postgres
  backends, so the subscription ceiling is much higher — this is the sanctioned driver
  for websocket-scale consumer counts.
- `go-redis` handles reconnect and resubscribe internally; the driver's loop only drains
  and forwards, dropping non-blocking to the consumer per the package-wide backpressure
  rule.
- Publish is a plain `PUBLISH` — **never transactional**. It fires immediately, even if a
  surrounding `data.Run` later rolls back. Apps following the out-of-band rule (publish
  after the commit) are unaffected; apps that relied on the Postgres driver's tx-coupling
  were leaning on a driver-specific extra, and a driver swap surfaces that.
- Channel scoping follows the Redis logical database in the URL, not the Postgres
  database. Environments sharing one Redis must separate by logical DB (`redis://host/1`)
  or by distinct instances.

Introspection: `PUBSUB CHANNELS` / `PUBSUB NUMSUB` on the Redis side, plus the
`pubsub listen` / `notify` CLI (which resolves `PUBSUB_URL` the same way application code
does).
