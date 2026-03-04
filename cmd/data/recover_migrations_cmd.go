package data

import (
	"fmt"
	"os"
	"path/filepath"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var recoverMigrationsCmd = &cobra.Command{
	Use:   "recover-migrations [output-dir]",
	Short: "Export migration cache from database to files",
	Run:   runRecoverMigrationsCmd,
}

func runRecoverMigrationsCmd(cmd *cobra.Command, args []string) {
	var (
		ctx, db = cmdutil.NewDataContext()
		prompt  = prompts.New(config.FromContext(ctx), args)
		outdir  = prompt.Str("output dir")
	)

	migrations, err := migrator.Load(migrator.FromDB(ctx, db))
	if err != nil {
		fxlog.Fatalf("recover-migrations: %w", err)
	}

	if len(migrations) == 0 {
		fxlog.Log("no migrations to recover")
		return
	}

	for _, migration := range migrations {
		fmt.Println(migration.Name)
	}
	fxlog.Log("migrations to recover", fxlog.Int("count", len(migrations)))
	if !prompt.YesNo("recover all migrations") {
		return
	}

	for _, migration := range migrations {
		upfile := filepath.Join(outdir, migration.Name+migrator.UpExt)
		fmt.Fprintln(os.Stdout, upfile)
		if err := os.WriteFile(upfile, []byte(migration.UpSQL), 0644); err != nil {
			fxlog.Fatalf("recover-migrations: %w", err)
		}

		downfile := filepath.Join(outdir, migration.Name+migrator.DownExt)
		fmt.Fprintln(os.Stdout, downfile)
		if err := os.WriteFile(downfile, []byte(migration.DownSQL), 0644); err != nil {
			fxlog.Fatalf("recover-migrations: %w", err)
		}
	}
}
