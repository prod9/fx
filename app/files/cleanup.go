package files

import (
	"context"
	"time"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/worker"
)

var (
	// UploadGracePeriodConfig controls how long a file row may exist without its
	// uploaded object before the sweep prunes it as an abandoned upload.
	UploadGracePeriodConfig = config.DurationDef("FILES_UPLOAD_GRACE_PERIOD", 24*time.Hour)

	// CleanupIntervalConfig controls the cadence of the files reconciliation sweep.
	CleanupIntervalConfig = config.DurationDef("FILES_CLEANUP_INTERVAL", 12*time.Hour)

	// CleanupBatchConfig controls the maximum number of file rows processed per sweep batch.
	CleanupBatchConfig = config.IntDef("FILES_CLEANUP_BATCH", 500)
)

// CleanupJob is the reconciliation worker. Register it with files.App (done) and seed the
// first run once at startup: worker.ScheduleNowIfNotExists(ctx, files.CleanupJob). The job
// reschedules itself thereafter.
var CleanupJob worker.Interface = &cleanupJob{}

type cleanupJob struct{}

func (j *cleanupJob) Name() string { return "files-cleanup" }

func (j *cleanupJob) Run(ctx context.Context) error {
	cfg := config.FromContext(ctx)
	interval := config.Get(cfg, CleanupIntervalConfig)
	gracePeriod := config.Get(cfg, UploadGracePeriodConfig)
	batchSize := config.Get(cfg, CleanupBatchConfig)

	defer worker.ScheduleInIfNotExists(ctx, CleanupJob, interval)
	return runCleanup(ctx, time.Now(), gracePeriod, batchSize)
}

// runCleanup prunes abandoned uploads: recent rows whose object never landed and that are
// older than the upload grace period. It is driven off the files table — each candidate
// row's object is probed directly — so the sweep only touches objects it owns and never
// enumerates the bucket. A row younger than the grace period is an upload still in flight
// and is left alone.
//
// The sweep never deletes objects; it only prunes rows. It bails on the first probe or
// delete error: every such error here is systemic (a dropped connection, a dead store), so
// there is nothing to salvage by continuing — the next run retries rather than hammering a
// broken dependency with the rest of the batch.
func runCleanup(
	ctx context.Context,
	now time.Time,
	gracePeriod time.Duration,
	batchSize int,
) error {
	return runCleanupWindow(ctx, FileFilter{
		CreatedAfter:  now.Add(-2 * gracePeriod),
		CreatedBefore: now.Add(-gracePeriod),
		Limit:         batchSize,
	})
}

func runFullCleanup(
	ctx context.Context,
	now time.Time,
	gracePeriod time.Duration,
	batchSize int,
) error {
	return runCleanupWindow(ctx, FileFilter{
		CreatedBefore: now.Add(-gracePeriod),
		Limit:         batchSize,
	})
}

func runCleanupWindow(ctx context.Context, filter FileFilter) error {
	for {
		files, err := SelectFiles(ctx, filter)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return nil
		}

		for _, f := range files {
			exists, err := blobstore.ObjectExists(ctx, f.RemotePath())
			if err != nil {
				return err
			}
			if exists {
				continue
			}

			key := FileKey{
				Kind:      f.Kind,
				OwnerType: f.OwnerType,
				OwnerID:   f.OwnerID,
				ID:        f.ID,
			}
			if _, err := DestroyFile(ctx, key); err != nil {
				return err
			}
		}

		filter.AfterID = files[len(files)-1].ID
	}
}
