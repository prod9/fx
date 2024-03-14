package data

import (
	"fmt"
	"path/filepath"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"
	"github.com/spf13/cobra"
)

var listMigrationsCmd = &cobra.Command{
	Use:   "list-migrations (outdir)",
	Short: "List all detected migration files.",
	RunE:  runListMigrationsCmd,
}

func runListMigrationsCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("list-migrations", &err)

	cfg := config.Configure()
	migrations, err := migrator.LoadAuto(cfg)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		upPath := filepath.Join(migration.Dir, migration.Name+migrator.UpExt)
		fmt.Println(upPath)
	}

	return nil
}
