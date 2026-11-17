package worker

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

var errJob = errors.New("job failed")

func TestProcessJob_DefaultFiltersRecoverPanics(t *testing.T) {
	w := New(fxtest.Configure(), &Stub{JobName: "panics", RunFunc: func(context.Context) error { panic("kaboom") }})

	var err error
	require.NotPanics(t, func() { err = w.processJob(testContext(), &Job{Name: "panics", Payload: "{}"}) })
	require.ErrorContains(t, err, "kaboom")
}

func TestProcessJob_ReturnsRunError(t *testing.T) {
	w := New(fxtest.Configure(), &Stub{JobName: "fails", RunFunc: func(context.Context) error { return errJob }})

	err := w.processJob(testContext(), &Job{Name: "fails", Payload: "{}"})
	require.ErrorIs(t, err, errJob)
}

func TestSetFilters_ReplacesDefaultsInOrder(t *testing.T) {
	var calls []string
	trace := func(label string) Filter {
		return func(_ context.Context, _ int64, job Interface) (Interface, error) {
			return Stub{JobName: job.Name(), RunFunc: func(ctx context.Context) error {
				calls = append(calls, label)
				return job.Run(ctx)
			}}, nil
		}
	}
	w := New(fxtest.Configure(), &Stub{JobName: "traced", RunFunc: func(context.Context) error {
		calls = append(calls, "run")
		return nil
	}})
	w.SetFilters(trace("first"), trace("second"))

	require.NoError(t, w.processJob(testContext(), &Job{Name: "traced", Payload: "{}"}))
	require.Equal(t, []string{"second", "first", "run"}, calls)
}

func TestProcessJob_FilterErrorFailsJob(t *testing.T) {
	ran := false
	w := New(fxtest.Configure(), &Stub{JobName: "blocked", RunFunc: func(context.Context) error {
		ran = true
		return nil
	}})
	w.SetFilters(func(context.Context, int64, Interface) (Interface, error) { return nil, errJob })

	err := w.processJob(testContext(), &Job{Name: "blocked", Payload: "{}"})
	require.ErrorIs(t, err, errJob)
	require.False(t, ran)
}

func testContext() context.Context {
	return config.NewContext(context.Background(), fxtest.Configure())
}

func TestLogFilter_LogsJobID(t *testing.T) {
	logs := captureLogs(t)
	w := New(fxtest.Configure(), &Stub{JobName: "logged", RunFunc: func(context.Context) error { return nil }})

	require.NoError(t, w.processJob(testContext(), &Job{ID: 42, Name: "logged", Payload: "{}"}))
	require.Contains(t, logs.String(), "msg=completed job=logged id=42")
}

func TestProcessJob_LogsPreparationFailure(t *testing.T) {
	logs := captureLogs(t)
	w := New(fxtest.Configure())

	err := w.processJob(testContext(), &Job{ID: 7, Name: "unregistered", Payload: "{}"})
	require.Error(t, err)
	require.Contains(t, logs.String(), "msg=failed job=unregistered id=7")
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	logs := &bytes.Buffer{}
	fxlog.SetSink(fxlog.NewSLogSink(slog.New(slog.NewTextHandler(logs, nil))))
	t.Cleanup(func() { fxlog.SetSink(nil) })
	return logs
}
