package main

import (
	"context"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/ctrlc"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/pubsub"
	"github.com/spf13/cobra"
)

var PongerCmd = &cobra.Command{
	Use:   "ponger",
	Short: "Echo every ping straight back as a pong",
	Run:   runPonger,
}

func runPonger(cmd *cobra.Command, args []string) {
	ctx, _ := cmdutil.NewDataContext()
	ctx, cancel := context.WithCancel(ctx)
	ctrlc.Do(cancel)

	pings, unsubscribe, err := pubsub.Subscribe(ctx, Ping)
	if err != nil {
		fxlog.Fatalf("pingpong ponger: %w", err)
	}
	defer unsubscribe()

	fxlog.Log("ponger ready")
	for {
		select {
		case ping := <-pings:
			fxlog.Log("pong", fxlog.Int("rally", ping.Rally))
			if err := pubsub.Publish(ctx, Pong, Ball{Rally: ping.Rally}); err != nil {
				fxlog.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
