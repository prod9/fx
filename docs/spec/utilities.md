# Utilities

**Status:** accepted

Small support packages that don't warrant their own spec. Each is independently
optional — import what you need.

* `fx.prodigy9.co/blobstore` — S3-compatible object storage client. Used by the
  `files` app fragment for presigned URL uploads and downloads. Public surface is
  presigned URLs + `DeleteObject` only; no server-side `Put`.
* `fx.prodigy9.co/cache` — In-memory and Redis caching with a unified interface.
* `fx.prodigy9.co/cmd/prompts` — Interactive TUI prompts for CLI commands (text
  input, list selection, yes/no confirmation). Inputs can be provided as positional
  args for scripting. Set `CI=1` for non-interactive mode, `ALWAYS_YES=1` to
  auto-confirm.
* `fx.prodigy9.co/ctrlc` — Graceful CTRL-C / SIGINT signal handler.
* `fx.prodigy9.co/passwords` — bcrypt password hashing.
* `fx.prodigy9.co/secret` — AES-256-GCM encryption for passing secrets safely
  (`Hide` / `Reveal`).
* `fx.prodigy9.co/validate` — Input validation helpers.
* `fx.prodigy9.co/slices` — Slice helpers.
