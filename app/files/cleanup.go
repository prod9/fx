package files

import (
	"context"
	"time"

	"fx.prodigy9.co/blobstore"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/worker"
)

// cleanupInterval is how often the cleanup sweep runs; deadTimeout is how long a file
// row may exist without its object uploaded before the sweep treats it as an abandoned
// upload and prunes it (uploads for large files can legitimately take a while).
const (
	cleanupInterval = 1 * time.Hour
	deadTimeout     = 24 * time.Hour
)

var filesCleanup worker.Interface = &cleanupJob{}

type cleanupJob struct{}

func (j *cleanupJob) Name() string { return "files.cleanup" }

func (j *cleanupJob) Run(ctx context.Context) error {
	defer worker.ScheduleInIfNotExists(ctx, filesCleanup, cleanupInterval)
	return runCleanup(ctx, time.Now())
}

// ScheduleCleanup seeds the first cleanup run. Call it once at app startup with a
// database-bearing context; the job reschedules itself thereafter. files.App registers
// the job on mount — this line starts it. Kept explicit (rather than auto-seeded on
// mount or gated behind config) so the schedule is visible and the consumer can wrap,
// guard, or re-time it however they need.
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

// planCleanup is the pure cleanup decision: an object with no owning row is an orphan to
// delete; a row whose object never arrived and is older than dead is an abandoned upload
// to prune. A young row with no object yet is an upload in progress and is left alone.
func planCleanup(rows []cleanupRow, storeKeys []string, now time.Time, dead time.Duration) cleanupPlan {
	rowByPath := make(map[string]cleanupRow, len(rows))
	for _, r := range rows {
		rowByPath[r.RemotePath] = r
	}
	inStore := make(map[string]bool, len(storeKeys))
	for _, k := range storeKeys {
		inStore[k] = true
	}

	var plan cleanupPlan
	for _, k := range storeKeys {
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

func runCleanup(ctx context.Context, now time.Time) error {
	storeKeys, err := blobstore.ListObjects(ctx, "")
	if err != nil {
		return err
	}

	files := []*File{}
	if err := data.Select(ctx, &files, `SELECT * FROM files`); err != nil {
		return err
	}

	rows := make([]cleanupRow, len(files))
	for i, f := range files {
		rows[i] = cleanupRow{ID: f.ID, RemotePath: f.RemotePath(), CreatedAt: f.CreatedAt}
	}

	plan := planCleanup(rows, storeKeys, now, deadTimeout)
	for _, key := range plan.ObjectsToDelete {
		if err := blobstore.DeleteObject(ctx, key); err != nil {
			fxlog.Log("files.cleanup: delete object failed",
				fxlog.String("key", key), fxlog.String("error", err.Error()))
		}
	}
	for _, id := range plan.RowsToPrune {
		if err := data.Exec(ctx, `DELETE FROM files WHERE id = $1`, id); err != nil {
			fxlog.Log("files.cleanup: prune row failed",
				fxlog.String("error", err.Error()))
		}
	}
	return nil
}
