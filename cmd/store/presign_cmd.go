package store

import (
	"context"
	"fmt"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var presignGetCmd = &cobra.Command{
	Use:   "presign-get <key>",
	Short: "Print a presigned GET URL for a key",
	Args:  cobra.ExactArgs(1),
	Run:   runPresignGetCmd,
}

var presignPutCmd = &cobra.Command{
	Use:   "presign-put <key>",
	Short: "Print a presigned PUT URL for a key",
	Args:  cobra.ExactArgs(1),
	Run:   runPresignPutCmd,
}

func runPresignGetCmd(cmd *cobra.Command, args []string) {
	url, err := blobstore.PresignedGetURL(context.Background(), args[0])
	if err != nil {
		fxlog.Fatalf("store presign-get: %w", err)
	}
	fmt.Println(url)
}

func runPresignPutCmd(cmd *cobra.Command, args []string) {
	url, err := blobstore.PresignedPutURL(context.Background(), args[0])
	if err != nil {
		fxlog.Fatalf("store presign-put: %w", err)
	}
	fmt.Println(url)
}
