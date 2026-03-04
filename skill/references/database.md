# Database, Transactions, and Migrations

Read this file when working with the `data` package, writing queries, managing
transactions, or creating/running migrations.

## Connection

Configure with `DATABASE_URL` env var. Connect via:

```go
db, err := data.Connect(cfg)
db := data.MustConnect(cfg)
ctx = data.NewContext(ctx, db)
```

Or use `middlewares.AddDataContext` (included via `AddDefaults()`) to inject `*sqlx.DB`
into request context automatically.

HA is handled externally (pgbouncer etc.), not in-app.

## Query Methods

All methods expect `*sqlx.DB` in context. Parameters: `ctx`, `dest`, `sql`, `args...`.

- `data.Get(ctx, dest, sql, args...)` — single row into struct
- `data.Select(ctx, dest, sql, args...)` — multiple rows into slice
- `data.Exec(ctx, sql, args...)` — no return value
- `data.Prepare(ctx, sql)` — prepared `*sqlx.Stmt`

Check `data.IsNoRows(err)` for not-found cases.

```go
func GetTodoByID(ctx context.Context, id string) (*Todo, error) {
	todo := &Todo{}
	if err := data.Get(ctx, todo, "SELECT * FROM todos WHERE id = $1", id); err != nil {
		return nil, err
	}
	return todo, nil
}
```

## Transactions

Transactions use `data.Scope` (an interface wrapping `*sqlx.Tx`). Scope provides `Get`,
`Select`, `Exec`, `Prepare` methods mirroring the top-level `data` functions.

### `data.Run` (recommended)

Closure-scoped transaction with auto commit/rollback:

```go
func CreateOrder(ctx context.Context, order *Order) error {
	return data.Run(ctx, func(scope data.Scope) error {
		err := scope.Exec("INSERT INTO orders ...")
		if err != nil {
			return err
		}
		// pass scope.Context() to propagate transaction to nested calls
		return CreateOrderItems(scope.Context(), order.Items)
	})
}
```

### `data.NewScopeErr` (flat style)

Returns cancel func to defer. Takes pointer to return error for commit/rollback decision:

```go
func GetUser(ctx context.Context, out any, id int64) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()

	err = scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	return GetUserProfile(scope.Context(), out, id)
}
```

### Manual `data.NewScope`

Full control — create scope and defer `scope.End(&err)`:

```go
func DoWork(ctx context.Context) (err error) {
	scope, err := data.NewScope(ctx, nil)
	if err != nil {
		return err
	}
	defer scope.End(&err)

	// use scope.Get, scope.Exec, etc.
	return nil
}
```

## Migrations

SQL file-based migration engine in `data/migrator`. Included via `AddDefaults()` or
manually with `Commands(data.Cmd)`.

### Embedding for production

```go
//go:embed **/*.sql
var sqlMigrations embed.FS

func main() {
	app.Build().
		AddDefaults().
		EmbedMigrations(sqlMigrations).
		Start()
}
```

Migrator checks embedded sources first, then current folder and parents.

### CLI commands

- `data migrate` — run all pending migrations
- `data rollback` — revert last migration
- `data new-migration (name) [subdir]` — create up+down SQL files
- `data list-migrations` — list detected migration files
- `data create-db` — create configured database
- `data psql` — open psql shell
- `data collect-migrations (outdir)` — collect files into one directory
- `data recover-migrations [outdir]` — export migration cache from DB
- `data resync-migrations` — sync DB cache with program files

### Scripting / CI

Set `CI=1` for non-interactive mode. Set `ALWAYS_YES=1` to auto-confirm prompts.
