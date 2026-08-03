package pubsub

import (
	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/fxlog"
	fxpubsub "fx.prodigy9.co/pubsub"
	"github.com/spf13/cobra"
)

var notifyCmd = &cobra.Command{
	Use:   "notify <channel> <payload>",
	Short: "Publish a payload to a channel",
	Args:  cobra.ExactArgs(2),
	Run:   runNotifyCmd,
}

func runNotifyCmd(cmd *cobra.Command, args []string) {
	ctx, _ := cmdutil.NewDataContext()

	if err := fxpubsub.PublishRaw(ctx, args[0], []byte(args[1])); err != nil {
		fxlog.Fatalf("pubsub notify: %w", err)
	}
	fxlog.Log("published", fxlog.String("channel", args[0]))
}
