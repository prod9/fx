# Fragment migrations example

Three child app fragments (`users`, `posts`, `comments`) each register their own
embedded migrations via `app.Builder.EmbedMigrations`. The root app in `main.go`
mounts all three but does not embed any migrations itself.

This exercises the aggregation behavior introduced in v0.8.5: `app.Start` walks the
child tree and registers each fragment's embedded FS with the migrator, which then
merges them and sorts by filename (timestamp-prefixed, so lex = chronological).

## Layout

```
examples/migrations/
  main.go                Root app — Mount(users.App, posts.App, comments.App)
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

The FK chain (`comments → posts → users`) means the merged migrations *must* apply in
filename order, or the FKs fail to resolve. That's the verification: if aggregation
preserves order across fragments, `data migrate` succeeds; if order is wrong,
Postgres rejects the FK.

## Running

```sh
cd examples/migrations
export DATABASE_URL=postgres://localhost/fx_migrations_example
export ALWAYS_YES=1                       # skip the apply-confirmation prompt

go run . data create-db
go run . data migrate                     # applies all 3 in order
go run . data rollback                    # repeat 3× to walk back down
dropdb fx_migrations_example              # cleanup (no built-in drop-db cmd)
```

Expected: `data migrate` reports three migrations applied. `\dt` in psql shows
`users`, `posts`, `comments`, plus the `migrations` bookkeeping table.

## What this does NOT exercise

`migrator.LoadAuto` prefers disk SQL over embedded if any `*.up.sql` lives at the
CWD. Running `go run .` from this directory works because there are no `.sql` files
at the top level — only inside the child packages. If you copy one of the child
`*.sql` files up to `examples/migrations/`, the disk-discovery branch fires and
embedded aggregation is bypassed entirely. See `docs/TODO.md` for the open question
about merging vs. exclusive precedence.
