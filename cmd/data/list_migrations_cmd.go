package data

import (
	"fmt"
	"path/filepath"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var listMigrationsCmd = &cobra.Command{
	Use:   "list-migrations",
	Short: "List all detected migration files.",
	Run:   runListMigrationsCmd,
}

func runListMigrationsCmd(cmd *cobra.Command, args []string) {
	_, cfg := cmdutil.NewBasicContext()
	migrations, err := migrator.LoadAuto(cfg)
	if err != nil {
		fxlog.Fatalf("list-migrations: %w", err)
	}

	for _, migration := range migrations {
		upPath := filepath.Join(migration.Dir, migration.Name+migrator.UpExt)
		downPath := filepath.Join(migration.Dir, migration.Name+migrator.DownExt)
		fmt.Println(upPath)
		fmt.Println(downPath)
	}
}
