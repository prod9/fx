package files

import (
	"time"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/config"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove file metadata for missing objects",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, _, cleanup := cmdutil.NewDataContext()
		defer cleanup()
		cfg := config.FromContext(ctx)
		gracePeriod := config.Get(cfg, UploadGracePeriodConfig)
		batchSize := config.Get(cfg, CleanupBatchConfig)

		return runFullCleanup(ctx, time.Now(), gracePeriod, batchSize)
	},
}
