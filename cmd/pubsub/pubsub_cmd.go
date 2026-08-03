package pubsub

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "pubsub",
	Short: "Work with Postgres LISTEN/NOTIFY pub/sub",
}

func init() {
	Cmd.AddCommand(
		notifyCmd,
		listenCmd,
	)
}
