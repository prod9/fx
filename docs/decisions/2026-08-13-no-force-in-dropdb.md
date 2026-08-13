# No `WITH (FORCE)` in data.DropDB

- **Date:** 2026-08-13
- **PR:** manual
- **Status:** accepted

## Decision

`data.DropDB` keeps plain `DROP DATABASE` — never `DROP DATABASE ... WITH (FORCE)`.

## Rationale

A drop failing with "database is being accessed by other users" is a deliberate
alarm: some test leaked a connection (a stray worker goroutine, an unclosed scope,
an unreleased pool). `WITH (FORCE)` would kill those connections and make the drop
succeed, masking the leak permanently — the same sin as raising a timeout to make a
slow test pass. The obvious default (FORCE, so test teardown "just works") is
exactly what this ruling rejects.

fxtest's own pool is already sequenced correctly: `t.Cleanup` runs LIFO, so
`db.Close()` happens before the drop. When a drop fails, the bug is in the
offending test's connection management — fix the test, don't soften the drop.

Context: this ruling came out of the 2026-08 connection-leak audit, where
`CreateDB`/`DropDB` leaking their own admin connections (fixed in `1cc8518`) had
been exhausting postgres `max_connections` and orphaning test databases.
