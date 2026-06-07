# Database

**Status:** accepted

The `data` package is completely optional. It mainly wraps `github.com/jmoiron/sqlx` to
make it a bit easier to control transaction scopes, properly handle rollbacks on
errors, etc.

Configure with the `DATABASE_URL` environment variable, usually just adds an entry in
`.env`:

```sh
DATABASE_URL=postgres:///mynewdb?sslmode=disable
```

Connect to the db with:

```go
// using DATABASE_URL env var
db, err := data.Connect(cfg)

// override with config.Set
config.Set(data.DatabaseURLConfig, "postgres:///mydb")
db, err := data.Connect(cfg)

// manually adding data middleware
err := app.Build().
	Name("my todo app").
	Middlewares(middlewares.AddDataContext()).
	// ...
	Start()
```

HA setups, if needed, are assumed to be handled outside the application at the
infrastructure level with something like `pgbouncer`.

Once set up, there are basic methods to interact with the database:

* `data.Get` — Executes a query and returns a single row.
* `data.Select` — Executes a query and returns multiple rows.
* `data.Exec` — Executes a query with no return value.
* `data.Run` — Runs a function inside a transaction (see Transactions section below).
* `data.Prepare` — Prepares a statement and returns a `*sqlx.Stmt` object.

Example:

```go
func GetTodoByID(ctx context.Context, id string) (*Todo, error) {
  const sql = "SELECT * FROM todos WHERE id = $1"

  todo := &Todo{}
  if err := data.Get(ctx, todo, sql, id); err != nil {
    return nil, err
  } else {
    return todo, nil
  }
}
```

Note that the `*sqlx.DB` is assumed to already be in the context. If not, you can use
`data.NewContext` like with the `config` package to pass it around:

```go
cfg := config.Configure()
db := data.MustConnect(cfg)

ctx := context.Background()
ctx = data.NewContext(ctx, db)
```

The methods usually take the following parameters, in order:

* `ctx` — Context, usually from the request.
* `dest` — Destination object to scan into. Arrays for `Select`, structs for `Get`.
* `sql` — SQL query to execute.
* `args` — Arguments to pass to the query.

There's an unfinished version of support for SQL Generators like `go-jet` with `GetSQL`
and similar. They should work, but largely untested, and very alpha.

## Transactions

Transactions are controlled through `data.Scope`. A scope wraps a `*sqlx.Tx` and
automatically issues `COMMIT` or `ROLLBACK` based on the returned error.

### `data.Run` (recommended)

The simplest way to run a transaction. The scope is created and ended automatically:

```go
func GetUserByID(ctx context.Context, out any, id int64) error {
	return data.Run(ctx, func(scope data.Scope) error {
		err := scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
		if err != nil {
			return err
		}

		// pass scope.Context() to propagate the transaction to nested calls
		var profile *UserProfile
		if err = GetUserProfile(scope.Context(), profile, id); err != nil {
			return err
		}

		user.Profile = profile
		return nil
	})
}
```

### `data.NewScopeErr` (flat style)

If you prefer a flat function body over a closure, `data.NewScopeErr` returns a cancel
function to `defer`. It takes a pointer to the return error so it can decide whether to
commit or rollback:

```go
func GetUserByID(ctx context.Context, out any, id int64) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	defer cancel()

	// watch out for the scope of the `err` variable
	err = scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	var profile *UserProfile
	if err = GetUserProfile(scope.Context(), profile, id); err != nil {
		return err
	}

	user.Profile = profile
	return nil
}
```

### Manual scope management

For full control, create a scope with `data.NewScope` and call `scope.End(&err)` in a
`defer`:

```go
func GetUserByID(ctx context.Context, out any, id int64) (err error) {
	var scope data.Scope
	scope, err = data.NewScope(ctx, nil)
	if err != nil {
		return err
	}
	defer scope.End(&err)

	err = scope.Get(out, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	var profile *UserProfile
	if err = GetUserProfile(scope.Context(), profile, id); err != nil {
		return err
	}

	user.Profile = profile
	return nil
}
```

### Notes

* A pointer to the return error is passed so the scope can automatically rollback on
  error.
* `data.Scope` is an interface wrapping `*sqlx.Tx` — it provides `Get`, `Select`,
  `Exec`, and `Prepare` methods mirroring the top-level `data` functions.
