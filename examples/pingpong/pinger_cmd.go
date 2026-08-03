package main

import (
	"context"
	"time"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/ctrlc"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/pubsub"
	"github.com/spf13/cobra"
)

const (
	// volleyPause paces the rally so it is watchable rather than a flood.
	volleyPause = 500 * time.Millisecond
	// serveAfter re-serves once the rally has gone this long without a pong — covering a
	// cold ponger at startup or a ball lost to the best-effort NOTIFY contract.
	serveAfter = 2 * time.Second
)

var PingerCmd = &cobra.Command{
	Use:   "pinger",
	Short: "Serve a ball and volley every pong back as the next ping",
	Run:   runPinger,
}

func runPinger(cmd *cobra.Command, args []string) {
	ctx, _ := cmdutil.NewDataContext()
	ctx, cancel := context.WithCancel(ctx)
	ctrlc.Do(cancel)

	pongs, unsubscribe, err := pubsub.Subscribe(ctx, Pong)
	if err != nil {
		fxlog.Fatalf("pingpong pinger: %w", err)
	}
	defer unsubscribe()

	ping(ctx, 1)
	idle := time.NewTicker(serveAfter)
	defer idle.Stop()

	for {
		select {
		case pong := <-pongs:
			idle.Reset(serveAfter)
			time.Sleep(volleyPause)
			ping(ctx, pong.Rally+1)
		case <-idle.C:
			fxlog.Log("no pong — re-serving")
			ping(ctx, 1)
		case <-ctx.Done():
			return
		}
	}
}

func ping(ctx context.Context, rally int) {
	fxlog.Log("ping", fxlog.Int("rally", rally))
	if err := pubsub.Publish(ctx, Ping, Ball{Rally: rally}); err != nil {
		fxlog.Error(err)
	}
}
