package data

import (
	"context"
	"fmt"
	"log"

	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"
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
		dir    = config.Get(cfg, data.MigrationPathConfig)
	)

	db, err := data.Connect(cfg)
	if err != nil {
		return err
	}

	var (
		ctx = data.NewContext(context.Background(), db)
		mig = migrator.New(db, dir)
	)

	plans, dirty, err := mig.Plan(ctx, migrator.IntentSync)
	if !dirty {
		log.Println("migrations are up-to-date")
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
	}
	log.Println(len(plans), "migrations to sync")
	if !prompt.YesNo("apply changes") {
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
		if err := mig.Apply(ctx, plan); err != nil {
			log.Fatalln("failed to run migration", err)
		}
	}

	log.Println(len(plans), "migration(s) synchronized")
	return nil
}
