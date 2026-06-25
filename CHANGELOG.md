# Changelog

## Unreleased

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
