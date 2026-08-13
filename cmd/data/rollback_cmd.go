package data

import (
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/fxlog"

	"github.com/spf13/cobra"
)

func buildRollbackCmd(srcs []migrator.Source) *cobra.Command {
	return &cobra.Command{
		Use:   "rollback",
		Short: "Revert one previously ran migration.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runMigration(migrator.IntentRollback, args, srcs); err != nil {
				fxlog.Fatalf("rollback: %w", err)
			}
		},
	}
}
