package cmd

import (
	"time"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/worker"
	"github.com/spf13/cobra"
)

var workerCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove stale job rows that need no further processing",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, cleanup := cmdutil.NewDataContext()
		defer cleanup()
		staleDuration := config.Get(config.FromContext(ctx), worker.StaleDurationConfig)

		return worker.Cleanup(ctx, time.Now(), staleDuration)
	},
}

func BuildWorkerCommand(jobs ...worker.Interface) *cobra.Command {
	workerCmd := &cobra.Command{
		Use:   "worker",
		Short: "Starts background worker.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Configure()
			w := worker.New(cfg, jobs...)
			return w.StartContext(cmd.Context())
		},
	}

	workerCmd.AddCommand(workerCleanupCmd)
	return workerCmd
}
