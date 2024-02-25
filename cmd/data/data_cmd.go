package data

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "data",
	Short: "Work with databases",
}

func init() {
	Cmd.AddCommand(
		collectMigrationsCmd,
		createDBCmd,
		migrateCmd,
		newMigrationCmd,
		psqlCmd,
		recoverMigrationCmd,
		rollbackCmd,
		syncMigrationsCmd,
	)
}
