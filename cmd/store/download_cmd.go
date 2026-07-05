package store

import (
	"context"
	"io"
	"net/http"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <key> [dest]",
	Short: "Download a key to a file (or stdout) via a presigned GET URL",
	Args:  cobra.RangeArgs(1, 2),
	Run:   runDownloadCmd,
}

func runDownloadCmd(cmd *cobra.Command, args []string) {
	key := args[0]

	url, err := blobstore.PresignedGetURL(context.Background(), key)
	if err != nil {
		fxlog.Fatalf("store download: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		fxlog.Fatalf("store download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fxlog.Fatalf("store download: unexpected status %s", resp.Status)
	}

	dest, closeDest, err := openDest(arg(args, 1))
	if err != nil {
		fxlog.Fatalf("store download: %w", err)
	}
	defer closeDest()

	if _, err := io.Copy(dest, resp.Body); err != nil {
		fxlog.Fatalf("store download: %w", err)
	}
}
