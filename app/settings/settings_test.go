package settings

import (
	"context"
	"testing"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// createSettingsTable applies the fragment's embedded migration against the test database,
// standing in for `data migrate` so the schema exists before the primitives run.
func createSettingsTable(t *testing.T, ctx context.Context) {
	t.Helper()
	up, err := migrations.ReadFile("202407301306_create_settings.up.sql")
	require.NoError(t, err)
	require.NoError(t, data.Exec(ctx, string(up)))
}

// TestGetFallback covers the default-on-absent contract: a missing key yields the caller's
// fallback (never an error), and once written the stored value wins.
func TestGetFallback(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	createSettingsTable(t, ctx)

	value, err := Get(ctx, "theme", "system")
	require.NoError(t, err, "an absent key must not be an error")
	require.Equal(t, "system", value, "an absent key must return the fallback")

	require.NoError(t, (&Upsert{Key: "theme", Value: "dark"}).Execute(ctx, &Settings{}))

	value, err = Get(ctx, "theme", "system")
	require.NoError(t, err)
	require.Equal(t, "dark", value, "a stored value must win over the fallback")
}
