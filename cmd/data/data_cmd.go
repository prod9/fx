package data

import (
	"fx.prodigy9.co/data/migrator"
	"github.com/spf13/cobra"
)

// Commands builds the data subcommands with the given migration sources threaded into
// every migration-reading command. Sources join LoadAuto's embedded tier, so a
// configured DATABASE_MIGRATIONS path or working-directory files still take precedence.
func Commands(srcs ...migrator.Source) []*cobra.Command {
	return []*cobra.Command{
		buildCollectMigrationsCmd(srcs),
		buildListMigrationsCmd(srcs),
		buildMigrateCmd(srcs),
		buildResyncMigrationsCmd(srcs),
		buildRollbackCmd(srcs),
		createDBCmd,
		newMigrationCmd,
		psqlCmd,
		recoverMigrationsCmd,
	}
}
