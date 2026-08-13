# Changelog

## v0.10.2

* **pubsub:** Pluggable internal drivers, selected by `PUBSUB_URL` scheme. Unset or
  `postgres://` keeps the existing LISTEN/NOTIFY backend; `redis://`/`rediss://` routes
  the same typed API over Redis Pub/Sub instead — useful when Postgres connection
  headroom is tight or a Redis is already in the topology. Public API is unchanged;
  there are no user-facing extension points. See `docs/spec/pubsub-redis.md`.
* **clients:** New `clients/redis` package — the single place FX dials Redis. Hands out
  one memoized `*redis.Client` per URL so pubsub, cache, and app code share a
  connection pool instead of each dialing their own.
* **cache:** The Redis cache now obtains its client from `clients/redis`; disconnecting
  a cache no longer closes the underlying client other subsystems may share.

* **data:** `CreateDB`/`DropDB` now close their admin connection. Previously each call
  leaked an idle connection to the default database for the life of the process, which
  exhausted postgres `max_connections` in test suites creating one database per test
  (via `fxtest.ConnectTestDatabase`) and orphaned test databases when the cleanup drop
  could no longer connect.
* **cmd:** `cmdutil.NewDataContext`/`NewMigratorContext` return a third value — a
  cleanup func that closes the connection pool. Callers defer it; one-shot CLI runs
  behave the same, embedded or repeated use no longer leaks a pool per call.

## v0.10.0

* **pubsub:** New `pubsub` package — a typed pub/sub built on Postgres LISTEN/NOTIFY.
  Declare a channel at package scope (`var OrdersChanged = pubsub.NewChannel[OrderEvent]
  ("orders_changed")`; `struct{}` for signal-only channels), then `Publish(ctx, ch,
  payload)` and `Subscribe(ctx, ch)` with JSON-encoded typed payloads; `PublishRaw`/
  `SubscribeRaw` expose the string floor. Subscriptions hold one pooled connection for
  their lifetime and auto-reconnect on a tight backoff. It is a latency optimization over
  polling, not a delivery guarantee — NOTIFY is at-most-once, so treat a notification as
  "something changed, look now" and re-derive from the table; anything needing guaranteed
  delivery uses `worker`. Channel names are whitelisted to plain identifiers (LISTEN can't
  parameterize its name) and payloads are capped at the 8000-byte NOTIFY ceiling. See
  `docs/spec/pubsub.md` and `examples/pubsub`.
* **data:** `DATABASE_MAX_OPEN` now defaults to 64 (was unlimited). A pubsub subscription
  holds a pooled connection for its whole life, so the pool needs a finite ceiling
  operators can size against Postgres `max_connections`; 64 leaves headroom under a
  default `max_connections` of 100. `DATABASE_MAX_IDLE` may now be 0.

## v0.9.2

* **app/settings:** New settings fragment — a key/value store table owned by an embedded
  migration (applied under `data migrate`, like the `audit` and `files` fragments).
  `settings.App` is a plain fragment var; `List`/`Get`/`Set`/`Delete` are the Go API, with
  `Get(ctx, key, fallback)` treating absence as the fallback rather than an error. A
  `config.Provider` (`settings.NewProvider`) backs config lookups with the same table.
  The REST controller is mounted deliberately, not auto-mounted: call `settings.MountRoutes`
  inside a route group you own and guard (`GET /settings`, `POST /settings/{slug}` upsert,
  `DELETE /settings/{slug}`), so settings data is never exposed by mounting `App` alone.
  Writes are upserts — the first write to a key creates its row.

## v0.9.1

* **fxtest:** New `FXTEST_SKIP_DBTESTS` flag — when truthy, any test calling
  `ConnectTestDatabase` skips instead of failing, so hermetic environments with no
  database (e.g. the in-build publish test gate) pass `go test ./...` cleanly. The
  repo's committed `.env` now sets it by default; re-enable DB tests locally with
  `FXTEST_SKIP_DBTESTS=0` in `.env.local` or the shell env.
* **fxtest:** *Breaking.* Renamed `TestDisableCleanup` to `CleanupConfig`, matching the
  `*Config` naming every other config var uses (the `FXTEST_CLEANUP` env name is
  unchanged).

## v0.9.0

The `files` changes below are breaking.

* **blobstore / blobserver / store:** New local disk blob server
  (`blobstore/blobserver`) exposing the S3 verbs `blobstore` uses — GET/PUT/DELETE/HEAD —
  over a local directory, so apps can develop against on-disk blobs before pointing at
  live S3. New `store` command group (`serve`, `upload`, `download`, `presign-get`,
  `presign-put`, `delete`). `blobstore` now selects transport from the `STORAGE_URL`
  scheme (`s3`/`https` TLS, `http` for the local server), adds `ObjectExists`, and fixes
  the misspelled `StorgeURLConfig` identifier to `StorageURLConfig` (the `STORAGE_URL` env
  name is unchanged).
* **files:** *Breaking.* Redesigned the controller API and made the fragment
  self-contained:
  - Removed `NewApp` and all `*blobstore.Client` threading (`WithClient`, the client
    params on `File` methods / `Destroy*` / `UploadInfoFromFile`). `files.App` is a plain
    fragment var; the store is reached through the global `blobstore` funcs
    (config-driven).
  - Removed `WithKind` (kind is the positional arg) and the entire link-age override
    (`WithLinkAge`, `resolveLinkAge`, `FILE_LINK_AGE`) — the presigned TTL is a fixed 1m
    package const.
  - Controllers now default to `ModeReadOnly`; writes require an explicit
    `WithMode(ModeReadWrite)` (fail-closed).
  - Added `Kind.MaxSize` (bytes; `0` ⇒ 256 MiB default, `-1` ⇒ unlimited), enforced in
    `CreateFile` with a coded field error before a presigned PUT is minted.
  - `files.App` now registers a `files-cleanup` worker that prunes abandoned uploads —
    rows whose object never landed and that are older than 24h — probing each candidate
    row's object directly rather than enumerating the bucket, so the sweep never deletes
    objects. Seed the first run once at startup with
    `worker.ScheduleNowIfNotExists(ctx, files.CleanupJob)`; it reschedules itself thereafter.
* **audit:** Document downstream ledger reconciliation for a service adopting the fragment
  in place of its own local audit migration — reset a disposable DB, or resync/recover the
  ledger so fx's migration records as already-applied.

## v0.8.7

* **audit:** New `app/audit` fragment — an append-only audit trail (`Record`/`Log`/`List`
  over an `audit_events` table) mounted like `settings.App`. Ported from TIES; actions and
  actors are caller-owned constants, and the read endpoint stays caller-side.

## v0.8.6

* **prompts:** Reimplemented on `golang.org/x/term` instead of `pterm`, dropping
  `pterm` and its subtree (~10 modules) from the dependency graph with nothing added
  (`x/term` was already transitive). The existing method API is unchanged; adds
  `Session.MultiSelect(question, defaults, options)` (arg order mirrors `List`) for
  multi-choice selection — `defaults` pre-checks entries, and the base method bails when
  non-interactive, with `OptionalMultiSelect` mirroring `OptionalList` for the
  skip-prompt case (space toggles, enter confirms). Interactive UI now renders to stderr
  so stdout stays clean for piping. TTY detection still uses `mattn/go-isatty`.
* **fxlog:** *Breaking.* `Sink` no longer requires `Fatal`. Process termination is
  now owned by package-level `fxlog.Fatal`: log the error via `Sink.Error`, flush
  the sink if it implements the new optional `Flusher{ Flush() error }`,
  `os.Stderr.Sync()`, then `os.Exit(1)`. `ZerologSink` no longer calls zerolog's
  inline `.Fatal()` builder. Custom `Sink` implementations must remove their
  `Fatal` method.

## v0.8.5

* **app, migrator:** Fragment migrations now aggregate through `Mount`. `app.Start`
  walks the child tree and registers each fragment's `EmbeddedMigrations()`, so
  fragments like `files.App` no longer require the root app to re-embed their
  migrations. `migrator.Embed` accumulates instead of replacing; `LoadAuto` merges
  registered sources and sorts by name (timestamp-prefixed filenames → chronological).
* **examples:** New `examples/migrations/` demonstrates fragment migration
  aggregation across three child fragments with an FK chain.
* **docs:** Split `DOCS.md` into per-topic specs under `docs/spec/`. New
  `docs/spec/releasing.md` covers the release process. `docs/{decisions,notes}/` and
  `docs/TODO.md` scaffolded for durable, point-in-time, and impermanent notes.

## v0.8.4

* **app/files:** New S3-backed file management package with presigned URLs and PostgreSQL metadata.
* **migrator:** Fix nil `*sqlx.DB` panic in `FromDB` migration source.
* **migrator:** Move migrations table bootstrap from `Plan` to `Apply`, making `Plan` read-only.
* **migrator:** `FromDB` checks table existence via `pg_tables` instead of mutating schema.

## v0.8.3

* **cmd:** `new-migration` now takes name as the first arg, subdirectory is optional second arg.
* **cmd:** Add `OptionalList` prompt variant for optional list selection with a default.
* **cmd:** Improve error handling across CLI commands.
* **data:** Rework `recover-migrations` command, extract `FromDB` as a migration source.
* **migrator:** Replace `IntentUpdate`/`IntentRecover` with `IntentResync`.
* **docs:** Document `CI=1` and `ALWAYS_YES=1` for scripting, update DOCS.md throughout.

## v0.8.2

* **prompts:** More robust TTY-interactivity detection using `go-isatty`.
* **cmd:** Fix help text arguments, messaging and naming.

## v0.8.1

* **worker:** Fix hot loop spinning on `signalIdled`.
* **worker:** Add indexes to jobs table for faster lookup.
* **data:** Add `DropDB` and `dbname` package for manipulating database names in URLs.
* **fxtest:** New package for test config and database helpers.
* **migrator:** Add tests.

## v0.8.0

* **fxlog:** New logging abstraction package (zerolog default, slog option).
* **docs:** Add comprehensive framework documentation (`DOCS.md`).
* Remove unmaintained `contrib/` directory.

## v0.7.0

Initial versioned release.
