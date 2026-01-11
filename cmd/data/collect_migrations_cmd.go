package data

import (
	"os"
	"path/filepath"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var collectMigrationsCmd = &cobra.Command{
	Use:   "collect-migrations (outdir)",
	Short: "Copies all detected migration files in the repository to the specified directory.",
	Run:   runCollectMigrationsCmd,
}

func runCollectMigrationsCmd(cmd *cobra.Command, args []string) {
	var (
		ctx, _ = cmdutil.NewDataContext()
		prompt = prompts.New(config.FromContext(ctx), args)
		outdir = prompt.Str("output dir")
	)

	migrations, err := migrator.LoadAuto(config.FromContext(ctx))
	if err != nil {
		fxlog.Fatalf("collect-migrations: %w", err)
	}

	for _, migration := range migrations {
		upPath := filepath.Join(outdir, migration.Name+migrator.UpExt)
		downPath := filepath.Join(outdir, migration.Name+migrator.DownExt)

		fxlog.Log("write", fxlog.String("path", upPath))
		if err := os.WriteFile(upPath, []byte(migration.UpSQL), 0644); err != nil {
			fxlog.Fatalf("collect-migrations: %w", err)
		}

		fxlog.Log("write", fxlog.String("path", downPath))
		if err := os.WriteFile(downPath, []byte(migration.DownSQL), 0644); err != nil {
			fxlog.Fatalf("collect-migrations: %w", err)
		}
	}
}
