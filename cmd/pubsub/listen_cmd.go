package pubsub

import (
	"context"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/ctrlc"
	"fx.prodigy9.co/fxlog"
	fxpubsub "fx.prodigy9.co/pubsub"
	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen <channel>...",
	Short: "Subscribe to channels and print notifications until interrupted",
	Args:  cobra.MinimumNArgs(1),
	Run:   runListenCmd,
}

func runListenCmd(cmd *cobra.Command, args []string) {
	ctx, _ := cmdutil.NewDataContext()

	cancels := make([]context.CancelFunc, 0, len(args))
	for _, name := range args {
		ch, cancel, err := fxpubsub.SubscribeRaw(ctx, name)
		if err != nil {
			fxlog.Fatalf("pubsub listen: %w", err)
		}
		cancels = append(cancels, cancel)
		go printNotifications(name, ch)
	}

	fxlog.Log("listening", fxlog.Any("channels", args))
	<-ctrlc.Chan()

	for _, cancel := range cancels {
		cancel()
	}
}

func printNotifications(name string, ch <-chan string) {
	for payload := range ch {
		fxlog.Log("notification",
			fxlog.String("channel", name),
			fxlog.String("payload", payload),
		)
	}
}
