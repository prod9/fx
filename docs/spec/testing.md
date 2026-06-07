# Testing

**Status:** accepted

The `fxtest` package provides helpers for writing tests against `fx` components.

* `fxtest.Configure()` — Returns a `*config.Source` initialized for testing (reads
  `.env` files, applies defaults).
* `fxtest.ConnectTestDatabase(t)` — Creates an isolated test database and returns a
  `context.Context` carrying both the config source and `*sqlx.DB`. The database is
  automatically dropped when the test completes, unless `FXTEST_CLEANUP=no` is set.

```go
func TestSomething(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	// ctx carries *config.Source and *sqlx.DB — pass to data functions directly
	err := data.Exec(ctx, "INSERT INTO todos (title) VALUES ($1)", "test")
	require.NoError(t, err)
}
```
