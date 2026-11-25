package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
)

// TestCleanup asserts the cleanup contract in docs/spec/workers.md "Cleanup": a running
// row past the stale duration is failed as abandoned; finished rows past the retention
// are deleted, except each job name's latest failure; pending rows are never touched.
func TestCleanup(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	require.NoError(t, ensureJobsTable(ctx))

	const stale = time.Hour
	var (
		now             = time.Now()
		pastRetention   = now.Add(-cleanupRetention - time.Hour)
		withinRetention = now.Add(-2 * stale)
		withinStale     = now.Add(-stale / 2)
	)

	olderFailure := insertJob(t, ctx, "flaky", FailedStatus, pastRetention.Add(-time.Hour))
	latestFailure := insertJob(t, ctx, "flaky", FailedStatus, pastRetention)
	oldCompleted := insertJob(t, ctx, "steady", CompletedStatus, pastRetention)
	recentCompleted := insertJob(t, ctx, "steady", CompletedStatus, withinRetention)
	oldPending := insertJob(t, ctx, "queued", PendingStatus, pastRetention)
	staleRunning := insertJob(t, ctx, "crashed", RunningStatus, withinRetention)
	liveRunning := insertJob(t, ctx, "busy", RunningStatus, withinStale)

	require.NoError(t, Cleanup(ctx, now, stale))

	require.False(t, jobExists(t, ctx, olderFailure), "superseded failure must be deleted")
	require.True(t, jobExists(t, ctx, latestFailure),
		"a job name's latest failure must survive retention")
	require.False(t, jobExists(t, ctx, oldCompleted), "completed job past retention must go")
	require.True(t, jobExists(t, ctx, recentCompleted),
		"completed job within retention must survive")
	require.True(t, jobExists(t, ctx, oldPending), "pending job must never be deleted")
	require.Equal(t, RunningStatus, jobStatus(t, ctx, liveRunning),
		"running job within the stale duration must be left alone")

	require.Equal(t, FailedStatus, jobStatus(t, ctx, staleRunning),
		"running job past the stale duration must be failed")
	require.Equal(t, abandonedError, jobError(t, ctx, staleRunning))
}

// TestCleanupJob_ReschedulesAtTwiceStale asserts the cadence in docs/spec/workers.md
// "Cleanup": the job runs again two stale durations after each run.
func TestCleanupJob_ReschedulesAtTwiceStale(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	stale := config.Get(config.FromContext(ctx), StaleDurationConfig)
	before := time.Now()

	require.NoError(t, CleanupJob.Run(ctx))

	next := Job{}
	require.NoError(t, data.Get(ctx, &next,
		`SELECT * FROM jobs WHERE name = $1 AND status = $2`,
		CleanupJob.Name(), PendingStatus))
	require.WithinRange(t, next.ScheduledAt,
		before.Add(2*stale), time.Now().Add(2*stale))
}

func insertJob(t *testing.T, ctx context.Context, name string, status JobStatus, updatedAt time.Time) int64 {
	id := int64(0)
	err := data.Get(ctx, &id, `
		INSERT INTO jobs (name, status, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id`,
		name, status, updatedAt)
	require.NoError(t, err)
	return id
}

func jobExists(t *testing.T, ctx context.Context, id int64) bool {
	exists := false
	err := data.Get(ctx, &exists, `SELECT EXISTS (SELECT 1 FROM jobs WHERE id = $1)`, id)
	require.NoError(t, err)
	return exists
}

func jobStatus(t *testing.T, ctx context.Context, id int64) JobStatus {
	status := JobStatus("")
	err := data.Get(ctx, &status, `SELECT status FROM jobs WHERE id = $1`, id)
	require.NoError(t, err)
	return status
}

func jobError(t *testing.T, ctx context.Context, id int64) string {
	errText := ""
	err := data.Get(ctx, &errText, `SELECT error FROM jobs WHERE id = $1`, id)
	require.NoError(t, err)
	return errText
}
