package data

import (
	"fmt"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"
	"fx.prodigy9.co/fxlog"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Runs all migration scripts in the configured migrations dir.",
	Run:   runMigrateCmd,
}

func runMigrateCmd(cmd *cobra.Command, args []string) {
	if err := runMigration(migrator.IntentMigrate, args); err != nil {
		fxlog.Fatalf("migrate: %w", err)
	}
}

func runMigration(intent migrator.Intent, args []string) (err error) {
	defer errutil.Wrap("migrate", &err)

	var (
		ctx, cfg = cmdutil.NewBasicContext()
		prompt   = prompts.New(cfg, args)
	)

	db, err := data.Connect(cfg)
	if err != nil {
		return err
	}

	migrator := migrator.New(db, migrator.FromAuto(cfg))
	plans, dirty, err := migrator.Plan(ctx, intent)
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
		if !prompt.YesNo("proceed with dirty migrations") {
			return nil
		}
	}

	fxlog.Log("migrations planned", fxlog.Int("migrations", len(plans)))
	if !prompt.YesNo("apply migrations") {
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
		if err = migrator.Apply(ctx, plan); err != nil {
			return err
		}
	}

	fxlog.Log("migration(s) applied", fxlog.Int("migrations", len(plans)))
	return nil
}
