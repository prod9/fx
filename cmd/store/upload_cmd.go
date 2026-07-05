package store

import (
	"context"
	"net/http"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <key> [src]",
	Short: "Upload a file (or stdin) to a key via a presigned PUT URL",
	Args:  cobra.RangeArgs(1, 2),
	Run:   runUploadCmd,
}

func runUploadCmd(cmd *cobra.Command, args []string) {
	key := args[0]
	body, size, closeSource, err := openSource(arg(args, 1))
	if err != nil {
		fxlog.Fatalf("store upload: %w", err)
	}
	defer closeSource()

	url, err := blobstore.PresignedPutURL(context.Background(), key)
	if err != nil {
		fxlog.Fatalf("store upload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		fxlog.Fatalf("store upload: %w", err)
	}
	req.ContentLength = size

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fxlog.Fatalf("store upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fxlog.Fatalf("store upload: unexpected status %s", resp.Status)
	}

	fxlog.Log("uploaded", fxlog.String("key", key))
}
