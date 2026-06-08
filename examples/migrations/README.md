# Fragment migrations example

Four child app fragments (`files`, `users`, `posts`, `comments`) each register their
own embedded migrations via `app.Builder.EmbedMigrations`. The root app in `main.go`
mounts all four but does not embed any migrations itself.

This exercises the aggregation behavior introduced in v0.8.5: `app.Start` walks the
child tree and registers each fragment's embedded FS with the migrator, which then
merges them and sorts by filename (timestamp-prefixed, so lex = chronological).

`files.App` is FX's own built-in fragment (`app/files`). Mounting it here proves the
same aggregation path picks up built-in fragments, not just user-defined ones — the
real-world case where an app pulls in `files.App`, `settings.App`, etc. and expects
their migrations to "just run."

## Layout

```
examples/migrations/
  main.go                Root app — Mount(files.App, users.App, posts.App, comments.App)
  users/
    users.go             EmbedMigrations(migrationsFS)
    20260601000001_create_users.{up,down}.sql
  posts/
    posts.go
    20260601000002_create_posts.{up,down}.sql   FK posts.user_id → users.id
  comments/
    comments.go
    20260601000003_create_comments.{up,down}.sql  FK comments.post_id → posts.id
```

`files.App` lives in `app/files/` and carries its own `202504041033_create_files`
migration — it sorts first lexicographically (2025 < 2026), independent of the
user/post/comment chain.

The FK chain (`comments → posts → users`) means the merged migrations *must* apply in
filename order, or the FKs fail to resolve. That's the verification: if aggregation
preserves order across fragments, `data migrate` succeeds; if order is wrong,
Postgres rejects the FK.

## Running

The example must be **built and run from a directory with no `*.up.sql` files**.
`migrator.LoadAuto` recursively scans CWD for SQL files and short-circuits on first
hit — running `go run .` from `examples/migrations/` finds the SQL files in the
child packages via disk-walk and never reaches the embedded path. That would only
demonstrate disk-mode discovery, not embedded aggregation. Build to a binary, run
from `/tmp` (or anywhere clean):

```sh
cd examples/migrations
export DATABASE_URL=postgres://localhost/fx_migrations_example
export ALWAYS_YES=1                       # skip the apply-confirmation prompt

go build -o /tmp/migrations-example .

/tmp/migrations-example data create-db
( cd /tmp && /tmp/migrations-example data migrate )    # applies all 4 in order
( cd /tmp && /tmp/migrations-example data rollback )   # repeat 4× to walk back down
dropdb fx_migrations_example              # cleanup (no built-in drop-db cmd)
```

Expected: `data migrate` reports four migrations applied, in this order:

```
migrate => 202504041033 create_files
migrate => 20260601000001 create_users
migrate => 20260601000002 create_posts
migrate => 20260601000003 create_comments
```

`\dt` in psql shows `files`, `users`, `posts`, `comments`, plus the `migrations`
bookkeeping table.

The disk-vs-embed precedence is itself a known footgun — see `docs/TODO.md`'s
"`migrator.LoadAuto`: merge disk + embed, or rethink loading" entry.

## What this does NOT exercise

`migrator.LoadAuto` prefers disk SQL over embedded if any `*.up.sql` lives at the
CWD. Running `go run .` from this directory works because there are no `.sql` files
at the top level — only inside the child packages. If you copy one of the child
`*.sql` files up to `examples/migrations/`, the disk-discovery branch fires and
embedded aggregation is bypassed entirely. See `docs/TODO.md` for the open question
about merging vs. exclusive precedence.
