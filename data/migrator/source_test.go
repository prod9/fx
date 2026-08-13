package migrator

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"fx.prodigy9.co/config"
	"github.com/stretchr/testify/require"
)

func TestCollect_UnionsAndSorts(t *testing.T) {
	migs, err := Collect(
		FromSQL("222_second", "SELECT 2", "SELECT 2"),
		FromSQL("111_first", "SELECT 1", "SELECT 1"),
	)

	require.NoError(t, err)
	require.Len(t, migs, 2)
	require.Equal(t, "111_first", migs[0].Name)
	require.Equal(t, "222_second", migs[1].Name)
}

func TestCollect_EmptyUnionIsNoMigrations(t *testing.T) {
	empty := Source(func() ([]Migration, error) { return nil, nil })
	noMigs := Source(func() ([]Migration, error) { return nil, ErrNoMigrations })

	_, err := Collect()
	require.ErrorIs(t, err, ErrNoMigrations)

	_, err = Collect(empty, noMigs)
	require.ErrorIs(t, err, ErrNoMigrations)
}

func TestCollect_SkipsNoMigrationsSourceInUnion(t *testing.T) {
	noMigs := Source(func() ([]Migration, error) { return nil, ErrNoMigrations })

	migs, err := Collect(noMigs, FromSQL("111_first", "SELECT 1", "SELECT 1"))

	require.NoError(t, err)
	require.Len(t, migs, 1)
	require.Equal(t, "111_first", migs[0].Name)
}

func TestCollect_PropagatesSourceError(t *testing.T) {
	boom := errors.New("boom")
	failing := Source(func() ([]Migration, error) { return nil, boom })

	_, err := Collect(failing)
	require.ErrorIs(t, err, boom)
}

func TestLoadAuto_ExtrasJoinEmbeddedTier(t *testing.T) {
	cfg := config.Configure()

	migs, err := LoadAuto(cfg, FromSQL("999_extra", "SELECT 1", "SELECT 1"))

	require.NoError(t, err)
	require.Len(t, migs, 1)
	require.Equal(t, "999_extra", migs[0].Name)
}

func TestLoadAuto_EnvPathOutranksExtras(t *testing.T) {
	dir := t.TempDir()
	writeMigrationPair(t, dir, "111_from_env")
	t.Setenv("DATABASE_MIGRATIONS", dir)
	cfg := config.Configure()

	migs, err := LoadAuto(cfg, FromSQL("999_extra", "SELECT 1", "SELECT 1"))

	require.NoError(t, err)
	require.Len(t, migs, 1)
	require.Equal(t, "111_from_env", migs[0].Name)
}

func writeMigrationPair(t *testing.T, dir, name string) {
	t.Helper()
	for _, ext := range []string{UpExt, DownExt} {
		path := filepath.Join(dir, name+ext)
		require.NoError(t, os.WriteFile(path, []byte("SELECT 1"), 0644))
	}
}
