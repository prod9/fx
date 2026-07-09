package files

import (
	"context"
	"time"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/worker"
)

// deadTimeout is how long a file row may exist without its uploaded object before the
// sweep prunes it as an abandoned upload. cleanupInterval is the sweep cadence.
const (
	deadTimeout     = 24 * time.Hour
	cleanupInterval = deadTimeout / 2
)

// CleanupJob is the reconciliation worker. Register it with files.App (done) and seed the
// first run once at startup: worker.ScheduleNowIfNotExists(ctx, files.CleanupJob). The job
// reschedules itself thereafter.
var CleanupJob worker.Interface = &cleanupJob{}

type cleanupJob struct{}

func (j *cleanupJob) Name() string { return "files-cleanup" }

func (j *cleanupJob) Run(ctx context.Context) error {
	defer worker.ScheduleInIfNotExists(ctx, CleanupJob, cleanupInterval)
	return runCleanup(ctx, time.Now())
}

// runCleanup prunes abandoned uploads: recent rows whose object never landed and that are
// older than deadTimeout. It is driven off the files table — each candidate row's object is
// probed directly — so the sweep only touches objects it owns and never enumerates the
// bucket. A row younger than deadTimeout is an upload still in flight and is left alone.
//
// The sweep never deletes objects; it only prunes rows. It bails on the first probe or
// delete error: every such error here is systemic (a dropped connection, a dead store), so
// there is nothing to salvage by continuing — the next run retries rather than hammering a
// broken dependency with the rest of the batch.
func runCleanup(ctx context.Context, now time.Time) error {
	cutoff := now.Add(-2 * deadTimeout)

	files := []*File{}
	if err := data.Select(ctx, &files, `SELECT * FROM files WHERE created_at > $1`, cutoff); err != nil {
		return err
	}

	for _, f := range files {
		if now.Sub(f.CreatedAt) <= deadTimeout {
			continue
		}

		exists, err := blobstore.ObjectExists(ctx, f.RemotePath())
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if err := data.Exec(ctx, `DELETE FROM files WHERE id = $1`, f.ID); err != nil {
			return err
		}
	}
	return nil
}
