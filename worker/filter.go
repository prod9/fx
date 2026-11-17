package worker

import (
	"context"
	"fmt"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/fxsentry"
)

// Filter runs after a job is prepared and before it runs. It returns the job, usually
// wrapped in a Stub whose Run defers around the inner Run. Filters apply in list order,
// each wrapping the result of the one before, so the last filter is outermost.
type Filter func(ctx context.Context, job Interface) (Interface, error)

// SetFilters replaces the worker's filters. New installs SentryFilter, RecoverFilter and
// LogFilter, in that order.
func (w *Worker) SetFilters(filters ...Filter) {
	w.Lock()
	defer w.Unlock()
	w.filters = filters
}

// SentryFilter scopes the run and reports its panic or returned error, when SENTRY_DSN
// is set. Place it before RecoverFilter so it sees panics first.
func SentryFilter(ctx context.Context, job Interface) (Interface, error) {
	if !fxsentry.Activate(config.FromContext(ctx)) {
		return job, nil
	}

	return Stub{JobName: job.Name(), RunFunc: func(runCtx context.Context) error {
		runCtx, scope := fxsentry.NewScope(runCtx)
		scope.SetTag("job", job.Name())
		defer fxsentry.CapturePanic(runCtx)

		err := job.Run(runCtx)
		if err != nil {
			fxsentry.Capture(runCtx, err)
		}
		return err
	}}, nil
}

// RecoverFilter turns a panic into the run's error, so the job is marked failed and the
// worker keeps running.
func RecoverFilter(_ context.Context, job Interface) (Interface, error) {
	return Stub{JobName: job.Name(), RunFunc: func(ctx context.Context) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("panic: %v", recovered)
			}
		}()
		return job.Run(ctx)
	}}, nil
}

// LogFilter logs the run's start, then its completion or failure with its duration.
func LogFilter(_ context.Context, job Interface) (Interface, error) {
	return Stub{JobName: job.Name(), RunFunc: func(ctx context.Context) error {
		fxlog.Log("running", fxlog.String("job", job.Name()))
		start := time.Now()

		err := job.Run(ctx)
		if err != nil {
			fxlog.Log("failed",
				fxlog.String("job", job.Name()),
				fxlog.Duration("duration", time.Since(start)),
				fxlog.Any("error", err),
			)
		} else {
			fxlog.Log("completed",
				fxlog.String("job", job.Name()),
				fxlog.Duration("duration", time.Since(start)),
			)
		}
		return err
	}}, nil
}
