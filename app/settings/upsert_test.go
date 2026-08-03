package settings

import (
	"context"
	"testing"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// TestSetUpserts drives Set against a fresh settings table: a first write to an unknown
// key must create the row (not fail on a missing UPDATE), and a second write to the same
// key must replace the value and advance updated_at while preserving created_at.
func TestSetUpserts(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	createSettingsTable(t, ctx)

	created, err := Set(ctx, "theme", "dark")
	require.NoError(t, err, "Set on an absent key must create the row")
	require.Equal(t, "theme", created.Key)
	require.Equal(t, "dark", created.Value)

	updated, err := Set(ctx, "theme", "light")
	require.NoError(t, err)
	require.Equal(t, "light", updated.Value, "second Set must replace the value")
	require.Equal(t, created.CreatedAt, updated.CreatedAt, "created_at must be preserved across updates")
	require.False(t, updated.UpdatedAt.Before(created.UpdatedAt), "updated_at must not move backwards")
}

func createSettingsTable(t *testing.T, ctx context.Context) {
	t.Helper()
	up, err := migrations.ReadFile("202407301306_create_settings.up.sql")
	require.NoError(t, err)
	require.NoError(t, data.Exec(ctx, string(up)))
}
