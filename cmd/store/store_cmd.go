package store

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "store",
	Short: "Work with blob storage",
}

func init() {
	Cmd.AddCommand(
		serveCmd,
		presignGetCmd,
		presignPutCmd,
		uploadCmd,
		downloadCmd,
		deleteCmd,
	)
}
