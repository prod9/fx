package cmd

import (
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/errutil"
	"fx.prodigy9.co/worker"
	"github.com/spf13/cobra"
)

func BuildWorkerCommand(jobs ...worker.Interface) *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Starts background worker.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errutil.Wrap("worker", &err)

			cfg := config.Configure()
			w := worker.New(cfg, jobs...)
			return w.Start()
		},
	}
}
