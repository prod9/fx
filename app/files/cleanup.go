package files

import (
	"context"
	"strconv"
	"strings"
	"time"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/worker"
)

// deadTimeout is how long a file row may exist without its uploaded object before the
// sweep prunes it as an abandoned upload. cleanupInterval is the sweep cadence.
const (
	deadTimeout     = 24 * time.Hour
	cleanupInterval = deadTimeout / 2
)

var filesCleanup worker.Interface = &cleanupJob{}

type cleanupJob struct{}

func (j *cleanupJob) Name() string { return "files-cleanup" }

func (j *cleanupJob) Run(ctx context.Context) error {
	defer worker.ScheduleInIfNotExists(ctx, filesCleanup, cleanupInterval)
	return runCleanup(ctx, time.Now())
}

// ScheduleCleanup seeds the first cleanup run; the job reschedules itself thereafter.
// Call once at startup with a database-bearing context.
func ScheduleCleanup(ctx context.Context) (int64, error) {
	return worker.ScheduleNowIfNotExists(ctx, filesCleanup)
}

// cleanupRow is the cleanup view of a files row: its object path and age.
type cleanupRow struct {
	ID         int64
	RemotePath string
	CreatedAt  time.Time
}

type cleanupPlan struct {
	ObjectsToDelete []string
	RowsToPrune     []int64
}

// planCleanup decides the sweep actions over a window of recent rows: an in-window object
// with no owning row is an orphan to delete; a row whose object never arrived and is older
// than dead is an abandoned upload to prune. A young row with no object yet is an upload
// still in progress and is left alone.
//
// The window floor is the lowest row id; objects with a smaller id predate the window and
// were settled by earlier sweeps, so they are skipped. With no rows there is no window and
// nothing is reconciled.
//
// files.App must own the STORAGE_URL bucket exclusively: any in-window object without a
// files row is deleted. A foreign object in the same bucket WILL be removed — give files
// its own bucket.
func planCleanup(rows []cleanupRow, storeKeys []string, now time.Time, dead time.Duration) cleanupPlan {
	if len(rows) == 0 {
		return cleanupPlan{}
	}

	idFloor := rows[0].ID
	rowByPath := make(map[string]cleanupRow, len(rows))
	for _, r := range rows {
		rowByPath[r.RemotePath] = r
		if r.ID < idFloor {
			idFloor = r.ID
		}
	}

	inStore := make(map[string]bool, len(storeKeys))
	var plan cleanupPlan
	for _, k := range storeKeys {
		if id, ok := idFromKey(k); !ok || id < idFloor {
			continue
		}
		inStore[k] = true
		if _, owned := rowByPath[k]; !owned {
			plan.ObjectsToDelete = append(plan.ObjectsToDelete, k)
		}
	}
	for _, r := range rows {
		if inStore[r.RemotePath] {
			continue
		}
		if now.Sub(r.CreatedAt) > dead {
			plan.RowsToPrune = append(plan.RowsToPrune, r.ID)
		}
	}
	return plan
}

// idFromKey extracts the trailing file id from an object key (owner_type/kind/owner_id/id).
// Keys not ending in an integer are ignored by the sweep.
func idFromKey(key string) (int64, bool) {
	idx := strings.LastIndex(key, "/")
	if idx < 0 {
		return 0, false
	}
	id, err := strconv.ParseInt(key[idx+1:], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func runCleanup(ctx context.Context, now time.Time) error {
	cutoff := now.Add(-2 * deadTimeout)

	files := []*File{}
	if err := data.Select(ctx, &files, `SELECT * FROM files WHERE created_at > $1`, cutoff); err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	rows := make([]cleanupRow, len(files))
	for i, f := range files {
		rows[i] = cleanupRow{ID: f.ID, RemotePath: f.RemotePath(), CreatedAt: f.CreatedAt}
	}

	storeKeys, err := blobstore.ListObjects(ctx, "")
	if err != nil {
		return err
	}

	plan := planCleanup(rows, storeKeys, now, deadTimeout)
	for _, key := range plan.ObjectsToDelete {
		if err := blobstore.DeleteObject(ctx, key); err != nil {
			fxlog.Log("files: storage object cleanup failed",
				fxlog.String("key", key), fxlog.String("error", err.Error()))
		}
	}
	for _, id := range plan.RowsToPrune {
		if err := data.Exec(ctx, `DELETE FROM files WHERE id = $1`, id); err != nil {
			fxlog.Log("files: file record cleanup failed",
				fxlog.String("error", err.Error()))
		}
	}
	return nil
}
