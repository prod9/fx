package data

import (
	"fx.prodigy9.co/data/migrator"

	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Revert one previously ran migration.",
	Run:   runRollbackCmd,
}

func runRollbackCmd(cmd *cobra.Command, args []string) {
	runMigration(migrator.IntentRollback, args)
}
