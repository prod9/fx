<!-- not spec/decision because: investigation only; no design chosen, the deployment
     target (Cloudflare Workers + D1) is settled but every FX-side ruling is still open -->

# SQLite / Cloudflare D1 support and worker job cleanup — investigation

Three questions were investigated on 2026-10-19:

1. What would it take for FX to support SQLite?
2. Should FX import an SQL abstraction library to get there?
3. How should rows of recurring worker jobs be cleaned up as they accumulate?

The purpose behind question 1 turned out to be **running FX apps inside Cloudflare
Workers against Cloudflare D1** (both settled by the maintainer). Most of this document is
about that target; the plain-SQLite findings are kept because D1 speaks SQLite's dialect.

Nothing in the codebase was changed. Claims about third-party systems (D1, syumai/workers,
TinyGo, DuckDB) come from web sources listed at the end. Several of them are secondary
sources and are marked "verify" where a design would depend on them.

## 1. Postgres coupling in FX today

### 1.1 Connection and admin operations

| Area                       | Postgres-specific behavior                                         |
|----------------------------|--------------------------------------------------------------------|
| `data.Connect`             | `sqlx.Open("pgx", url)` is hard-coded                              |
| `data.CreateDB` / `DropDB` | Admin connection to the `postgres` database, `CREATE/DROP DATABASE` |
| `data/dbname`              | URL path is the database name; default database is `postgres`     |
| `cmd data psql`            | Shells out to `psql`                                               |
| `fxtest`                   | Creates and drops a database per test run via `CreateDB`/`DropDB`  |

For plain SQLite these become a driver chosen by `DATABASE_URL` scheme (for example
`modernc.org/sqlite`, pure Go, no cgo), file create/remove, and a `sqlite3` shell.

### 1.2 Table definitions (DDL)

Every table FX defines needs a SQLite-dialect variant:

| Location                              | Postgres-only construct                     | SQLite equivalent                 |
|---------------------------------------|---------------------------------------------|-----------------------------------|
| `worker/jobs.go` `CreateJobsTableSQL` | `SERIAL`, `TIMESTAMPTZ`                     | `INTEGER PRIMARY KEY`, see §1.4   |
| `app/files` migration                 | `SERIAL`, `TIMESTAMPTZ`                     | same                              |
| `app/audit` migration                 | `BIGSERIAL`, `JSONB`, `DEFAULT now()`       | `INTEGER PRIMARY KEY`, `TEXT`, `CURRENT_TIMESTAMP` |
| `app/settings` migration              | `TIMESTAMPTZ`                               | see §1.4                          |
| `data/migrator` `CheckMigrationsTableSQL` | `pg_tables` lookup                      | `sqlite_master`                   |
| `cmd/data/new_migration_cmd.go` template | `SERIAL`, `TIMESTAMPTZ`                  | per-dialect template              |

`SERIAL` is not merely cosmetic in SQLite: only a column declared exactly
`INTEGER PRIMARY KEY` becomes the auto-assigned rowid alias, so `id SERIAL NOT NULL
PRIMARY KEY` leaves `id` NULL and the insert fails.

Embedded fragment migrations (`app/files`, `app/audit`, `app/settings`) are plain SQL
files, so supporting SQLite means each fragment ships a migration set per dialect, or the
fragment is documented as Postgres-only.

### 1.3 Queries (DML)

Portable as-is:

- `RETURNING *` (SQLite ≥ 3.35), `INSERT … ON CONFLICT … DO UPDATE` (≥ 3.24)
- `NULLIF`, `RANDOM()`, `LIMIT`/`OFFSET`, the CTE used by `data/page.Select`
- `$N` placeholders, including reuse of the same `$N`. SQLite treats `$NNN` as a named
  parameter, numbered by first appearance, and maps repeats to the same slot. Verify with
  one test against the chosen driver.

Breaks:

- `app/audit` `Record`: the `$6::jsonb` cast is a syntax error in SQLite.
- Every timestamp comparison and ordering (see §1.4):
  - `worker` `FindPendingJobSQL`: `scheduled_at < CURRENT_TIMESTAMP`
  - `ORDER BY created_at` in `app/files` and `app/settings`
  - `ORDER BY occurred_at` in `app/audit`
  - the `app/files` cleanup grace-period comparison

Postgres-only by nature: `pubsub`'s Postgres driver (`pg_notify`, `LISTEN`/`UNLISTEN`).

SQL written by apps is the app's own responsibility; FX cannot make arbitrary SQL portable.

### 1.4 Timestamps are the real design decision

SQLite has no time type: timestamps are stored as text (or numbers) and compared as such.
The Go driver writes a `time.Time` argument in its own text format, which includes a zone
offset, while a `DEFAULT CURRENT_TIMESTAMP` column writes `YYYY-MM-DD HH:MM:SS` in UTC.
Rows written each way do not compare or sort correctly against each other, so queries
like the worker's pending-job lookup return wrong results without any error.

The two candidate conventions:

- Every timestamp is written from Go, and SQL-side defaults are dropped.
- Every timestamp is stored in one fixed representation (Unix epoch, or one fixed RFC 3339
  form) on both dialects.

Whether scanning a `TIMESTAMPTZ`-declared column back into `time.Time` works also depends
on the driver's declared-type parsing; verify per driver.

## 2. Transactions: SQLite vs Postgres

FX's transaction surface is small. `data.Scope` calls `BeginTxx(ctx, nil)` (default
options), then commits or rolls back in `End`. A nested scope joins the parent's
transaction; there are no savepoints, no `SELECT … FOR UPDATE`, and no explicit isolation
level anywhere in FX.

| Concern                    | Postgres                                   | SQLite                                                                  | Effect on FX                                                         |
|----------------------------|--------------------------------------------|-------------------------------------------------------------------------|----------------------------------------------------------------------|
| Concurrent writers         | Row-level locks; many writers              | One writer per database file; others wait on `busy_timeout` or get `SQLITE_BUSY` | Server and worker processes serialize all writes. WAL mode and `busy_timeout` are required in the DSN. |
| Lock upgrade               | Not an issue                               | Default `BEGIN` is DEFERRED; read-then-write can fail with `SQLITE_BUSY` without waiting | `takeOnePendingJob` (SELECT then UPDATE) hits this. Write transactions need `BEGIN IMMEDIATE`, which `database/sql` cannot express, so it goes in the DSN (`_txlock=immediate`; name varies per driver, verify for modernc). |
| Isolation                  | READ COMMITTED by default                  | Effectively SERIALIZABLE because writers are serialized                 | Stricter, nothing breaks. The scheduling race noted in `ScheduleJobIfNotExistsSQL` cannot happen. |
| Error inside a transaction | Any failed statement aborts the transaction | Usually only the failed statement is undone; `BUSY`/`FULL`/`NOMEM` roll back the whole transaction | Code that catches a DB error and continues inside a scope works on SQLite and fails on Postgres. FX does not do this; apps might. |
| DDL inside a transaction   | Supported, except e.g. `CREATE INDEX CONCURRENTLY` | Supported, but `PRAGMA foreign_keys` / `journal_mode` have no effect inside one | The migrator wraps each migration in a scope. SQLite table rebuilds (needed for most `ALTER`s) usually need `PRAGMA foreign_keys=off`, which then does nothing. |
| Context cancellation       | Supported                                  | Supported (modernc interrupts the running statement)                   | None.                                                                |

For plain SQLite the work is DSN defaults (WAL, `busy_timeout`, immediate locking) plus a
note about table-rebuild migrations. No `Scope` redesign is needed.

## 3. Cloudflare D1 specifics

D1 speaks SQLite's dialect, so §1 applies, but it differs from a SQLite file in two ways
that matter more than any dialect issue:

- **No SQL transactions (verify).** Secondary sources consistently report that D1 rejects
  `BEGIN`/`COMMIT`/`ROLLBACK`/`SAVEPOINT` and offers atomic batches (`db.batch()`)
  instead. A failed statement stops the batch. If this holds, `data.Run`/`data.Scope`
  cannot work as designed, and neither can the worker's job claim, which reads a row and
  conditionally updates it inside one transaction. A batch cannot express "read, decide in
  Go, then write". The design question becomes what `Scope` means on D1.
- **No wire protocol.** D1 is reachable only through a Workers binding or Cloudflare's
  HTTP API. Go `database/sql` drivers:
  - `github.com/mdmourao/go-d1` — for ordinary Go servers, beta, goes through a proxy
    Worker that the user deploys.
  - `syumai/workers` `cloudflare/d1` — opened with `sql.Open("d1", "<binding name>")`,
    usable only from Go code running inside Workers. Marked alpha. Whether it implements
    `BeginTx` (and what it does) is unverified.

### 3.1 DuckDB as a front for D1 — not viable

No DuckDB extension for D1 was found. DuckDB's sqlite extension attaches local SQLite
files only, and DuckDB cannot write to databases reached over HTTPS or S3. A front layer
would also inherit D1's missing transactions and per-query HTTP round trip, DuckDB is an
in-process analytics engine rather than an OLTP store, and its Go binding requires cgo.
DuckDB's Postgres-like dialect would not remove the need for SQLite-dialect statements,
since D1 still executes SQLite underneath.

## 4. Running FX inside Cloudflare Workers

Workers are not JS-only: Go compiles to Wasm. `syumai/workers` (and its newer module
`syumai/workers-go`, which requires Go ≥ 1.24) serves a plain `http.Handler` through
`workers.Serve(handler)`, and supports D1 (alpha), Cron Triggers, and Queues.

How each FX piece fares:

| FX piece                                       | On Workers                                                                                   |
|------------------------------------------------|----------------------------------------------------------------------------------------------|
| `httpserver/controllers`, `middlewares`, `render`, chi | Work; served via `workers.Serve(handler)` instead of `ListenAndServe`                 |
| `httpserver`, `cmd` (cobra)                    | Do not apply: there is no process to start and no CLI                                        |
| `config`                                       | Reads env vars and `.env` files; Workers have no filesystem and pass config as bindings, so a Workers config source is needed |
| `data` (sqlx)                                  | sqlx can wrap the D1 `database/sql` driver; `Scope` breaks without transactions (§3)         |
| `data/migrator`                                | Cannot run inside a Worker; runs outside, or is replaced by `wrangler d1 migrations`         |
| `worker`                                       | No background process for the poll loop; a Cron Trigger could drain the `jobs` table each tick |
| `pubsub`, `cache` (Redis), `blobstore` (minio) | No Postgres and no raw TCP from Go Wasm; outbound HTTP must go through `fetch`; R2 is the natural blob store |

### 4.1 Build-size and TinyGo constraint

- Workers scripts are capped at 3 MB (free plan) / 10 MB (paid plan), compressed.
- Standard Go Wasm binaries are large; `syumai/workers` examples all use TinyGo for this
  reason.
- TinyGo has reflection gaps. One D1-backed Go Workers app switched from `encoding/json`
  to easyjson because of them.
- FX depends on reflection: sqlx struct scanning, and `encoding/json` in `render` and in
  worker job payloads.

The gating questions, in order:

1. Does a standard-Go `GOOS=js GOARCH=wasm` build of a small FX app (for example
   `examples/todoapi`) fit under 10 MB compressed? Not yet measured.
2. If not, do sqlx and `encoding/json` work under TinyGo?

## 5. Importing an SQL abstraction library

Recommendation so far: import none. Not ruled on.

`sqlx` already is the abstraction layer (`DriverName()`, `Rebind()`). The alternatives
solve a different problem:

- Query builders (`goqu`, `squirrel`) abstract DML, which is the part that is already
  mostly portable. They do not touch DDL, migrations, admin commands, timestamps, or
  transactions, which is where the breakage is.
- ORMs (`bun`, `gorm`) replace sqlx instead of wrapping it, contrary to FX philosophy
  rule 2 ("thin wrappers over standard primitives, never replacements"), and would force
  every downstream app onto them.

On D1 the argument strengthens: no library restores transactions that D1 does not have.

## 6. Worker job row cleanup

Resolved: shipped as `worker.CleanupJob`; the current behavior is in
`docs/spec/workers.md` "Cleanup". What follows is the survey made before the decision.

`worker` has no cron primitive. A recurring job reschedules itself on each run
(`app/files` `CleanupJob` uses `ScheduleInIfNotExists`; `examples/workers` uses
`ScheduleIn`), so every run leaves one `completed` row in `jobs` forever. Nothing prunes
the table, and there is no index on `updated_at`.

| Option                              | Shape                                                                                                         | Trade-off                                                     |
|-------------------------------------|---------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------|
| A. `worker.PruneJob` (recommended)  | Self-rescheduling job; batched `DELETE … WHERE status = 'completed' AND updated_at < now - WORKER_RETENTION`; plus a `worker prune` command for operators | Same pattern as `files.CleanupJob`; opt-in (philosophy rule 5); operable in prod (rule 6) |
| B. Prune inside the worker loop     | The worker deletes old rows on each idle tick                                                                  | No registration step, but hidden destructive behavior in the poll loop |
| C. Reuse the row for recurring jobs | Rescheduling flips the completed row back to `pending`                                                         | Loses run history (against append-over-update); needs a "recurring job" concept the API lacks |

Retention: completed rows from routine runs lose their value once past a window, so
deleting them is justified. Failed rows are debugging evidence and should get a longer
retention, or none. The prune query wants an index on `(status, updated_at)`.

On Workers (§4) the poll loop disappears, so cleanup would ride the same Cron Trigger that
drains jobs; the retention rules above carry over unchanged.

## 7. Open questions for the follow-up session

1. Does a standard-Go Wasm build of a small FX app fit the 10 MB Workers limit (§4.1)?
2. If not, is TinyGo viable for sqlx and `encoding/json`?
3. What does `data.Scope` mean on D1, given no transactions? Does the worker's job claim
   get redesigned around a single conditional `UPDATE … RETURNING`?
4. Does the syumai D1 driver implement `BeginTx`, and how?
5. Timestamp convention across dialects (§1.4).
6. Per-dialect migrations for `app/*` fragments, or Postgres-only fragments?
7. How does `config` read Workers bindings?
8. Is plain-file SQLite (outside Workers) also a goal, or only D1?
9. Abstraction library: confirm "none" (§5).

## Sources

- [mdmourao/go-d1](https://pkg.go.dev/github.com/mdmourao/go-d1)
- [syumai/workers](https://github.com/syumai/workers) ·
  [syumai/workers-go](https://pkg.go.dev/github.com/syumai/workers-go) ·
  [DeepWiki: syumai/workers D1](https://deepwiki.com/syumai/workers/3.3-d1-database)
- [cxjava/workers README mirror (size limits)](https://github.com/cxjava/workers)
- [AdjectiveAllison/tasks (TinyGo + D1, easyjson)](https://github.com/AdjectiveAllison/tasks)
- [Oskar Dudycz: D1 transactions and trade-offs](https://event-driven.io/en/cloudflare_d1_transactions_and_tradeoffs/)
- [drizzle-orm issue #2463](https://github.com/drizzle-team/drizzle-orm/issues/2463)
- [DuckDB SQLite extension](https://duckdb.org/docs/lts/core_extensions/sqlite) ·
  [DuckDB over HTTPS/S3](https://duckdb.org/docs/current/guides/network_cloud_storage/duckdb_over_https_or_s3)
