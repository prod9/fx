package data

import (
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/dbname"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// CreateDB and DropDB open an admin connection to the default database; that
// connection must not outlive the call. Leaked admin pools accumulate idle
// connections for the life of the process, which exhausts postgres
// max_connections in test suites that create one database per test.
func TestCreateDropDBLeavesNoAdminConnections(t *testing.T) {
	baseURL := config.Get(config.Configure(), DatabaseURLConfig)
	if baseURL == "" {
		t.Skip("DATABASE_URL unset; skipping postgres-backed test")
	}
	adminURL, err := dbname.SetDefaultDB(baseURL)
	require.NoError(t, err)
	control, err := sqlx.Connect("pgx", adminURL)
	if err != nil {
		t.Skipf("postgres unreachable: %s", err)
	}
	defer control.Close()

	testURL, err := dbname.Set(baseURL, "fx_data_admin_conn_test")
	require.NoError(t, err)
	t.Setenv("DATABASE_URL", testURL)
	cfg := config.Configure()

	before := countAdminConnections(t, control)
	require.NoError(t, CreateDB(cfg))
	require.NoError(t, DropDB(cfg))
	after := countAdminConnections(t, control)

	require.Equal(t, before, after,
		"CreateDB/DropDB must close their admin connections")
}

func countAdminConnections(t *testing.T, control *sqlx.DB) int {
	count := 0
	err := control.Get(&count,
		`SELECT count(*) FROM pg_stat_activity
		 WHERE usename = current_user AND datname = current_database()
		   AND pid <> pg_backend_pid()`)
	require.NoError(t, err)
	return count
}
