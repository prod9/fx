package cmd

import (
	"io"
	"os"
	"testing"

	"fx.prodigy9.co/data/migrator"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestBuildDataCommand_HasAllSubcommands(t *testing.T) {
	cmd := BuildDataCommand()

	names := map[string]bool{}
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range []string{
		"collect-migrations",
		"create-db",
		"list-migrations",
		"migrate",
		"new-migration",
		"psql",
		"recover-migrations",
		"resync-migrations",
		"rollback",
	} {
		require.True(t, names[name], "missing subcommand: %s", name)
	}
}

func TestBuildDataCommand_ThreadsSources(t *testing.T) {
	cmd := BuildDataCommand(migrator.FromSQL("777_extra", "SELECT 1", "SELECT 1"))
	list := findSubcommand(t, cmd, "list-migrations")

	out := captureStdout(t, func() { list.Run(list, nil) })

	require.Contains(t, out, "777_extra")
}

func findSubcommand(t *testing.T, cmd *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, sub := range cmd.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	t.Fatalf("subcommand not found: %s", name)
	return nil
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	read, write, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = orig }()

	fn()

	require.NoError(t, write.Close())
	bytes, err := io.ReadAll(read)
	require.NoError(t, err)
	return string(bytes)
}
