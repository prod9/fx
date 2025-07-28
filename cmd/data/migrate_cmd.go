package data

import (
	"context"
	"fmt"

	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"
	"fx.prodigy9.co/fxlog"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [migrations-dir]",
	Short: "Runs all migration scripts in the configured migrations dir.",
	RunE:  runMigrateCmd,
}

func runMigrateCmd(cmd *cobra.Command, args []string) error {
	return runMigration(migrator.IntentMigrate, args)
}

func runMigration(intent migrator.Intent, args []string) (err error) {
	defer errutil.Wrap("migrate", &err)

	var (
		cfg    = config.Configure()
		prompt = prompts.New(cfg, args)
	)

	db, err := data.Connect(cfg)
	if err != nil {
		return err
	}

	scope, err := data.NewScope(context.Background(), db)
	if err != nil {
		fxlog.Fatalf("db connection failed: %w", err)
	} else {
		defer scope.End(&err)
	}

	migrator := migrator.New(db, migrator.FromAuto(cfg))
	plans, dirty, err := migrator.Plan(scope.Context(), intent)
	if err != nil {
		return err
	}

	if len(plans) == 0 {
		fxlog.Log("no changes")
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
	}

	if dirty {
		fxlog.Log("migrations are missing or have changed content")
		if !prompt.YesNo("proceed anyway") {
			return nil
		}
	}

	fxlog.Log("migrations planned", fxlog.Int("migrations", len(plans)))
	if !prompt.YesNo("apply changes") {
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
		if err = migrator.Apply(scope.Context(), plan); err != nil {
			fxlog.Fatalf("migration failed: %w", err)
			return
		}
	}

	fxlog.Log("migration(s) applied", fxlog.Int("migrations", len(plans)))
	return nil
}
