package worker

import (
	"context"
	"errors"
	"testing"

	"fx.prodigy9.co/config"
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
		return func(_ context.Context, job Interface) (Interface, error) {
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
	w.SetFilters(func(context.Context, Interface) (Interface, error) { return nil, errJob })

	err := w.processJob(testContext(), &Job{Name: "blocked", Payload: "{}"})
	require.ErrorIs(t, err, errJob)
	require.False(t, ran)
}

func testContext() context.Context {
	return config.NewContext(context.Background(), fxtest.Configure())
}
