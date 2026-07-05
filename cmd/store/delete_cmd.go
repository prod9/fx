package store

import (
	"context"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete an object",
	Args:  cobra.ExactArgs(1),
	Run:   runDeleteCmd,
}

func runDeleteCmd(cmd *cobra.Command, args []string) {
	key := args[0]
	if err := blobstore.DeleteObject(context.Background(), key); err != nil {
		fxlog.Fatalf("store delete: %w", err)
	}
	fxlog.Log("deleted", fxlog.String("key", key))
}
