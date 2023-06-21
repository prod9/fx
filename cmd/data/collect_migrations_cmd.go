package data

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"
	"github.com/spf13/cobra"
)

var collectMigrationsCmd = &cobra.Command{
	Use:   "collect-migrations (outdir)",
	Short: "Copies all detected migration files in the repository to the specified directory.",
	RunE:  runCollectMigrationsCmd,
}

func runCollectMigrationsCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("collect-migrations", &err)

	var (
		cfg    = config.Configure()
		prompt = prompts.New(cfg, args)
		outdir = prompt.Str("output dir")
		indir  = config.Get(cfg, data.MigrationPathConfig)
	)

	migrations, err := migrator.LoadMigrations(indir)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		var (
			upPath      = filepath.Join(migration.Dir, migration.Name+migrator.UpExt)
			downPath    = filepath.Join(migration.Dir, migration.Name+migrator.DownExt)
			outUpPath   = filepath.Join(outdir, migration.Name+migrator.UpExt)
			outDownPath = filepath.Join(outdir, migration.Name+migrator.DownExt)
		)

		if err := copyFile(outUpPath, upPath); err != nil {
			return err
		}
		log.Println(upPath, "=>", outUpPath)

		if err := copyFile(outDownPath, downPath); err != nil {
			return err
		}
		log.Println(downPath, "=>", outDownPath)
	}

	return nil
}

func copyFile(dest, src string) error {
	srcfile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcfile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()

	_, err = io.Copy(destfile, srcfile)
	return err
}
