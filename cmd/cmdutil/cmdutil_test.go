package cmdutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The cleanup returned by NewDataContext must release the connection pool;
// CLI commands defer it so embedded/multi-command use doesn't leak pools.
func TestNewDataContextCleanupClosesPool(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres:///fx_cmdutil_test?sslmode=disable")

	_, db, cleanup := NewDataContext()
	cleanup()

	require.ErrorContains(t, db.Ping(), "closed")
}
