package settings

import (
	"testing"

	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// TestUpsert drives the Upsert action against a bare database — no migration run, no table
// created up front — so it also proves the self-init guard: a first write to an unknown key
// creates the row, and a second write to the same key replaces the value and advances
// updated_at while preserving created_at.
func TestUpsert(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	created := &Settings{}
	require.NoError(t, (&Upsert{Key: "theme", Value: "dark"}).Execute(ctx, created),
		"upsert on a bare database must self-init the table and create the row")
	require.Equal(t, "theme", created.Key)
	require.Equal(t, "dark", created.Value)

	updated := &Settings{}
	require.NoError(t, (&Upsert{Key: "theme", Value: "light"}).Execute(ctx, updated))
	require.Equal(t, "light", updated.Value, "second upsert must replace the value")
	require.Equal(t, created.CreatedAt, updated.CreatedAt, "created_at must be preserved across updates")
	require.False(t, updated.UpdatedAt.Before(created.UpdatedAt), "updated_at must not move backwards")
}
