package store

import (
	"fx.prodigy9.co/blobstore/blobserver"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve blobs from local disk (S3-compatible, for local development)",
	Run:   runServeCmd,
}

func runServeCmd(cmd *cobra.Command, args []string) {
	cfg := config.Configure()
	if err := blobserver.New(cfg).Start(); err != nil {
		fxlog.Fatalf("store serve: %w", err)
	}
}
