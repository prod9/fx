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

var syncMigrationsCmd = &cobra.Command{
	Use:   "sync-migrations",
	Short: "sync migrations",
	RunE:  runSyncMigrationsCmd,
}

func runSyncMigrationsCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("sync-migrate", &err)

	var (
		cfg    = config.Configure()
		prompt = prompts.New(cfg, args)
	)

	db, err := data.Connect(cfg)
	if err != nil {
		return err
	}

	var (
		ctx = data.NewContext(context.Background(), db)
		mig = migrator.New(db, migrator.FromAuto(cfg))
	)

	plans, dirty, err := mig.Plan(ctx, migrator.IntentSync)
	if !dirty {
		fxlog.Log("migrations up-to-date")
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
	}
	fxlog.Log("migrations changed", fxlog.Int("migrations", len(plans)))
	if !prompt.YesNo("apply changes") {
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
		if err := mig.Apply(ctx, plan); err != nil {
			fxlog.Fatalf("sync failed: %w", err)
			return err
		}
	}

	fxlog.Log("migration(s) synchronized", fxlog.Int("migrations", len(plans)))
	return nil
}
