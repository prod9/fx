package worker

import (
	"context"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
)

var (
	// StaleDurationConfig controls how long a finished job row is kept before cleanup
	// deletes it, and how long a running job may go without finishing before cleanup
	// marks it abandoned. It is also the cleanup job's cadence.
	StaleDurationConfig = config.DurationDef("WORKER_STALE_DURATION", 48*time.Hour)

	// CleanupJob prunes stale job rows. Register it with the worker and seed the first run
	// once at startup: worker.ScheduleNowIfNotExists(ctx, worker.CleanupJob). The job
	// reschedules itself thereafter.
	CleanupJob Interface = &cleanupJob{}

	abandonedError = "abandoned"
)

type cleanupJob struct{}

func (j *cleanupJob) Name() string { return "worker-cleanup" }

func (j *cleanupJob) Run(ctx context.Context) error {
	staleDuration := config.Get(config.FromContext(ctx), StaleDurationConfig)

	defer ScheduleInIfNotExists(ctx, CleanupJob, staleDuration)
	return Cleanup(ctx, time.Now(), staleDuration)
}

// Cleanup removes job rows that need no further processing once they are older than
// staleDuration. It first fails running jobs whose worker vanished, then deletes stale
// completed and failed rows.
func Cleanup(ctx context.Context, now time.Time, staleDuration time.Duration) error {
	if err := ensureJobsTable(ctx); err != nil {
		return err
	}
	staleBefore := now.Add(-staleDuration)

	return data.Run(ctx, func(s data.Scope) error {
		if err := s.Exec(AbandonStaleJobsSQL,
			FailedStatus, abandonedError, now,
			RunningStatus, staleBefore,
		); err != nil {
			return err
		} else {
			return s.Exec(DeleteStaleJobsSQL,
				CompletedStatus, FailedStatus, staleBefore)
		}
	})
}
