package data

import (
	"fmt"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var resyncMigrationsCmd = &cobra.Command{
	Use:   "resync-migrations",
	Short: "Update database migration cache to match program files",
	Run:   runResyncMigrationsCmd,
}

var forceResync bool

func init() {
	resyncMigrationsCmd.Flags().BoolVar(&forceResync, "force", false, "")
	resyncMigrationsCmd.Flags().Lookup("force").Hidden = true
}

func runResyncMigrationsCmd(cmd *cobra.Command, args []string) {
	var (
		ctx, mig = cmdutil.NewMigratorContext()
		prompt   = prompts.New(config.FromContext(ctx), args)
	)

	plans, dirty, err := mig.Plan(ctx, migrator.IntentResync)
	if err != nil {
		fxlog.Fatalf("resync-migrations: %w", err)
		return
	} else if !dirty {
		fxlog.Log("migrations up-to-date")
		return
	}

	// assert(dirty)
	for _, plan := range plans {
		fmt.Println(plan)
	}
	fxlog.Log("migrations changed", fxlog.Int("migrations", len(plans)))

	if !forceResync {
		fxlog.Fatalf("dangerous operation, --force is required")
	}
	if !prompt.YesNo("re-synchronize migration content") {
		return
	}

	for _, plan := range plans {
		fmt.Println(plan)
		if err := mig.Apply(ctx, plan); err != nil {
			fxlog.Fatalf("resync-migrations: %w", err)
		}
	}

	fxlog.Log("migration(s) synchronized", fxlog.Int("migrations", len(plans)))
}
