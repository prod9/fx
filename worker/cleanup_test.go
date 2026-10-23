package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxtest"
)

// TestCleanup asserts the cleanup contract in docs/spec/workers.md "Cleanup": rows that
// need no further processing are deleted once stale, pending rows are never touched, and
// a running row past the stale duration is failed as abandoned and kept for one more
// stale period before it goes too.
func TestCleanup(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)
	require.NoError(t, ensureJobsTable(ctx))

	const stale = 48 * time.Hour
	var (
		now     = time.Now()
		old     = now.Add(-stale - time.Hour)
		recent  = now.Add(-time.Hour)
		nextRun = now.Add(stale + time.Minute)
	)

	oldCompleted := insertJob(t, ctx, CompletedStatus, old)
	oldFailed := insertJob(t, ctx, FailedStatus, old)
	oldPending := insertJob(t, ctx, PendingStatus, old)
	oldRunning := insertJob(t, ctx, RunningStatus, old)
	recentCompleted := insertJob(t, ctx, CompletedStatus, recent)
	recentFailed := insertJob(t, ctx, FailedStatus, recent)
	recentRunning := insertJob(t, ctx, RunningStatus, recent)

	require.NoError(t, Cleanup(ctx, now, stale))

	require.False(t, jobExists(t, ctx, oldCompleted), "stale completed job must be deleted")
	require.False(t, jobExists(t, ctx, oldFailed), "stale failed job must be deleted")
	require.True(t, jobExists(t, ctx, oldPending), "pending job must never be deleted")
	require.True(t, jobExists(t, ctx, recentCompleted), "recent completed job must survive")
	require.True(t, jobExists(t, ctx, recentFailed), "recent failed job must survive")
	require.Equal(t, RunningStatus, jobStatus(t, ctx, recentRunning),
		"recently running job must be left alone")

	require.Equal(t, FailedStatus, jobStatus(t, ctx, oldRunning),
		"stale running job must be failed as abandoned")
	require.True(t, jobExists(t, ctx, oldRunning),
		"abandoned job must stay visible until it goes stale as failed")

	require.NoError(t, Cleanup(ctx, nextRun, stale))
	require.False(t, jobExists(t, ctx, oldRunning),
		"abandoned job must be deleted once stale as failed")
}

func insertJob(t *testing.T, ctx context.Context, status JobStatus, updatedAt time.Time) int64 {
	id := int64(0)
	err := data.Get(ctx, &id, `
		INSERT INTO jobs (name, status, updated_at)
		VALUES ('cleanup-test', $1, $2)
		RETURNING id`,
		status, updatedAt)
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
