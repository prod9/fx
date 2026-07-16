package fxtest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectTestDatabaseSkipsWhenSkipDBTestsSet(t *testing.T) {
	t.Setenv("FXTEST_SKIP_DBTESTS", "1")

	skipped, reached := false, false
	t.Run("subject", func(st *testing.T) {
		defer func() { skipped = st.Skipped() }()
		ConnectTestDatabase(st)
		reached = true
	})

	require.True(t, skipped, "ConnectTestDatabase should skip the test")
	require.False(t, reached, "no code after ConnectTestDatabase should run")
}
