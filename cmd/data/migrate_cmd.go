package data

import (
	"context"
	"fmt"
	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"
	"log"

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
		dir    = config.Get(cfg, data.MigrationPathConfig)
	)

	db, err := data.Connect(cfg)
	if err != nil {
		return err
	}

	scope, err := data.NewScope(context.Background(), db)
	if err != nil {
		log.Fatalln("db connection error", err)
	} else {
		defer scope.End(&err)
	}

	migrator := migrator.New(db, dir)
	plans, dirty, err := migrator.Plan(scope.Context(), intent)
	if err != nil {
		return err
	}

	if len(plans) == 0 {
		log.Println("no changes")
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
	}

	if dirty {
		log.Println("some migrations are missing or have changed content")
		if !prompt.YesNo("proceed anyway") {
			return nil
		}
	}

	log.Println(len(plans), "migrations planned")
	if !prompt.YesNo("apply changes") {
		return nil
	}

	for _, plan := range plans {
		fmt.Println(plan)
		if err := migrator.Apply(scope.Context(), plan); err != nil {
			log.Fatalln("failed to run migration", err)
		}
	}

	log.Println(len(plans), "migration(s) applied")
	return nil
}
