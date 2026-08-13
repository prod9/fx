package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/dbname"
	"fx.prodigy9.co/fxtest"
)

func TestStartClosesDatabasePool(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	cfg := config.FromContext(ctx)

	name, err := dbname.From(config.Get(cfg, data.DatabaseURLConfig))
	require.NoError(t, err)
	baseline := countConnections(t, ctx, name)

	w := New(cfg)
	done := make(chan error, 1)
	go func() { done <- w.Start() }()

	require.Eventually(t, func() bool {
		w.Lock()
		defer w.Unlock()
		return w.cancel != nil
	}, 5*time.Second, 10*time.Millisecond, "worker should start")

	w.Stop()
	require.ErrorIs(t, <-done, ErrStop)

	require.Eventually(t, func() bool {
		return countConnections(t, ctx, name) <= baseline
	}, 5*time.Second, 50*time.Millisecond,
		"worker pool connections should be released after Start returns")
}

func TestScheduleIfNotExistsUnderContention(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	require.NoError(t, ensureJobsTable(ctx))

	const attempts = 20
	var (
		wg      sync.WaitGroup
		barrier = make(chan struct{})
		results = make(chan error, attempts)
	)
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-barrier
			_, err := ScheduleNowIfNotExists(ctx, &TestJob{})
			results <- err
		}()
	}
	close(barrier)
	wg.Wait()
	close(results)

	scheduled, rejected := 0, 0
	for err := range results {
		switch {
		case err == nil:
			scheduled++
		case errors.Is(err, ErrJobExists):
			rejected++
		default:
			require.NoError(t, err)
		}
	}

	require.Equal(t, 1, scheduled, "exactly one scheduler should win")
	require.Equal(t, attempts-1, rejected, "the rest should get ErrJobExists")
	require.Equal(t, 1, countPendingJobs(t, ctx, (&TestJob{}).Name()))
}

func countPendingJobs(t *testing.T, ctx context.Context, name string) int {
	count := 0
	err := data.Get(ctx, &count,
		`SELECT COUNT(*) FROM jobs WHERE name = $1 AND status = 'pending'`, name)
	require.NoError(t, err)
	return count
}

func countConnections(t *testing.T, ctx context.Context, dbName string) int {
	count := 0
	err := data.Get(ctx, &count,
		`SELECT COUNT(*) FROM pg_stat_activity WHERE datname = $1`, dbName)
	require.NoError(t, err)
	return count
}
