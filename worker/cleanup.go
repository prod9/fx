package worker

import (
	"context"
	"errors"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
)

var (
	// StaleDurationConfig is how long a job may stay running before cleanup marks it
	// failed as abandoned. A job running past it warrants an operator's look. Cleanup
	// runs every two stale durations.
	StaleDurationConfig = config.DurationDef("WORKER_STALE_DURATION", time.Hour)

	// CleanupJob prunes stale job rows. Register it with the worker and seed the first run
	// once at startup: worker.ScheduleNowIfNotExists(ctx, worker.CleanupJob). The job
	// reschedules itself thereafter.
	CleanupJob Interface = &cleanupJob{}
)

const (
	abandonedError = "abandoned"

	// cleanupRetention is how long finished rows are kept, apart from each job name's
	// latest failure, which is kept until a newer failure replaces it.
	cleanupRetention = 24 * time.Hour
)

type cleanupJob struct{}

func (j *cleanupJob) Name() string { return "worker-cleanup" }

func (j *cleanupJob) Run(ctx context.Context) (err error) {
	staleDuration := config.Get(config.FromContext(ctx), StaleDurationConfig)

	defer func() {
		_, schedErr := ScheduleInIfNotExists(ctx, CleanupJob, 2*staleDuration)
		if schedErr != nil && !errors.Is(schedErr, ErrJobExists) {
			err = errors.Join(err, schedErr)
		}
	}()
	return Cleanup(ctx, time.Now(), staleDuration)
}

// Cleanup fails running jobs older than staleDuration as abandoned, then deletes finished
// rows older than the retention, keeping each job name's latest failure.
func Cleanup(ctx context.Context, now time.Time, staleDuration time.Duration) error {
	if err := ensureJobsTable(ctx); err != nil {
		return err
	}

	return data.Run(ctx, func(s data.Scope) error {
		if err := s.Exec(AbandonStaleJobsSQL,
			FailedStatus, abandonedError, now,
			RunningStatus, now.Add(-staleDuration),
		); err != nil {
			return err
		}
		return s.Exec(DeleteStaleJobsSQL,
			CompletedStatus, FailedStatus, now.Add(-cleanupRetention))
	})
}
