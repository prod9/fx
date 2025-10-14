package migrator

import (
	"context"
	"testing"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

const (
	TestMigrationName    = "nothing"
	TestMigrationUpSQL   = "SELECT 1; --up"
	TestMigrationDownSQL = "SELECT 1; --down"
)

func TestMigrator_Schema(t *testing.T) {
	mig, _ := buildTestMigrator(t)

	plan, dirt, err := mig.Plan(t.Context(), IntentMigrate)
	require.NoError(t, err)
	require.False(t, dirt)
	require.Len(t, plan, 1)
}

func TestMigrator_Apply_Basic(t *testing.T) {
	mig, ctx := buildTestMigrator(t)

	plan, dirt, err := mig.Plan(t.Context(), IntentMigrate)
	require.NoError(t, err)
	require.False(t, dirt)
	require.Len(t, plan, 1)

	err = mig.Apply(t.Context(), plan[0])
	require.NoError(t, err)

	n := 0
	err = data.Get(ctx, &n, "SELECT COUNT(*) FROM migrations")
	require.NoError(t, err)
	require.Equal(t, 1, n)

	m := Migration{}
	err = data.Get(ctx, &m, "SELECT * FROM migrations WHERE name = $1", "nothing")
	require.NoError(t, err)

	t.Logf("%+v", m)
	require.Equal(t, TestMigrationName, m.Name)
	require.Equal(t, TestMigrationUpSQL, m.UpSQL)
	require.Equal(t, TestMigrationDownSQL, m.DownSQL)
}

func buildTestMigrator(t *testing.T) (*Migrator, context.Context) {
	ctx := fxtest.ConnectTestDatabase(t)

	mig := New(
		data.FromContext(ctx),
		FromSQL(TestMigrationName, TestMigrationUpSQL, TestMigrationDownSQL),
	)
	return mig, ctx
}
