package data

import (
	"fmt"
	"os"
	"path/filepath"

	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"
	"github.com/spf13/cobra"
)

var recoverMigrationCmd = &cobra.Command{
	Use:   "recover-migration [name]",
	Short: "recover migration",
	RunE:  runRecoverMigrationCmd,
}

func runRecoverMigrationCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("migrate", &err)

	var (
		cfg    = config.Configure()
		prompt = prompts.New(cfg, args)
		dir    = config.Get(cfg, migrator.MigrationPathConfig)
	)

	db, err := data.Connect(cfg)
	if err != nil {
		return err
	}

	migrations, err := migrator.RecoverMigrations(db)
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return fmt.Errorf("no migrations to recover")
	}

	recovered := prompts.GenList(
		prompt,
		"which migration to recover?",
		migrations[len(migrations)-1],
		migrations,
		func(m migrator.Migration) string { return m.Name },
	)

	fmt.Println(migrator.Plan{
		Action:    migrator.ActionRecover,
		Migration: recovered,
	})
	if !prompt.YesNo("recover migration") {
		return nil
	}

	upfile := filepath.Join(dir, recovered.Name+migrator.UpExt)
	fmt.Fprintln(os.Stdout, upfile)
	if err := os.WriteFile(upfile, []byte(recovered.UpSQL), 0644); err != nil {
		return err
	}

	downfile := filepath.Join(dir, recovered.Name+migrator.DownExt)
	fmt.Fprintln(os.Stdout, downfile)
	if err := os.WriteFile(downfile, []byte(recovered.DownSQL), 0644); err != nil {
		return err
	}

	return nil
}
