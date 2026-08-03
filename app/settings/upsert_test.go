package settings

import (
	"context"
	"testing"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// TestUpsert drives the Upsert action against a fresh settings table: a first write to an
// unknown key must create the row (not fail on a missing UPDATE), and a second write to
// the same key must replace the value and advance updated_at while preserving created_at.
func TestUpsert(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	createSettingsTable(t, ctx)

	created := &Settings{}
	require.NoError(t, (&Upsert{Key: "theme", Value: "dark"}).Execute(ctx, created),
		"upsert on an absent key must create the row")
	require.Equal(t, "theme", created.Key)
	require.Equal(t, "dark", created.Value)

	updated := &Settings{}
	require.NoError(t, (&Upsert{Key: "theme", Value: "light"}).Execute(ctx, updated))
	require.Equal(t, "light", updated.Value, "second upsert must replace the value")
	require.Equal(t, created.CreatedAt, updated.CreatedAt, "created_at must be preserved across updates")
	require.False(t, updated.UpdatedAt.Before(created.UpdatedAt), "updated_at must not move backwards")
}

func createSettingsTable(t *testing.T, ctx context.Context) {
	t.Helper()
	up, err := migrations.ReadFile("202407301306_create_settings.up.sql")
	require.NoError(t, err)
	require.NoError(t, data.Exec(ctx, string(up)))
}
