package settings

import (
	"testing"

	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// TestGetFallback covers the default-on-absent contract against a bare database: a missing
// key yields the caller's fallback (never an error), and once written the stored value wins.
func TestGetFallback(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	value, err := Get(ctx, "theme", "system")
	require.NoError(t, err, "an absent key must not be an error")
	require.Equal(t, "system", value, "an absent key must return the fallback")

	require.NoError(t, (&Upsert{Key: "theme", Value: "dark"}).Execute(ctx, &Settings{}))

	value, err = Get(ctx, "theme", "system")
	require.NoError(t, err)
	require.Equal(t, "dark", value, "a stored value must win over the fallback")
}
