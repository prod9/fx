package data

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "data",
	Short: "Work with databases",
}

func init() {
	Cmd.AddCommand(
		createDBCmd,
		migrateCmd,
		newMigrationCmd,
		rollbackCmd,
		collectMigrationsCmd,
	)
}
