package worker

import (
	"context"
	"time"

	"fx.prodigy9.co/data"
)

const (
	CreateJobsTableSQL = `
		CREATE TABLE IF NOT EXISTS jobs (
			id      SERIAL NOT NULL PRIMARY KEY,
			name    TEXT NOT NULL,
			status  TEXT NOT NULL DEFAULT 'pending',
			payload TEXT NOT NULL DEFAULT '',
			error   TEXT NOT NULL DEFAULT '',
			
			created_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP),
			scheduled_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP)
		);

		CREATE INDEX IF NOT EXISTS idx_jobs_name_status ON jobs(name, status);
		CREATE INDEX IF NOT EXISTS idx_jobs_status_scheduled_at ON jobs(status, scheduled_at);
		`

	ScheduleJobSQL = `
		INSERT INTO jobs (name, status, payload, scheduled_at)
		VALUES ($1, $2, $3, $4)
		RETURNING *;`

	// Insert-if-no-pending fused into one statement so two schedulers racing the same
	// name cannot both pass a separate lookup. A sliver of a race remains under READ
	// COMMITTED (two statements executing at the same instant see the same snapshot);
	// accepted over a partial unique index, which would forbid legitimate duplicate
	// pending jobs scheduled deliberately via ScheduleAt.
	ScheduleJobIfNotExistsSQL = `
		INSERT INTO jobs (name, status, payload, scheduled_at)
		SELECT $1, $2, $3, $4
		WHERE NOT EXISTS (
			SELECT 1 FROM jobs
			WHERE name = $1
				AND status = $2
		)
		RETURNING *;`

	// we could use FOR UPDATE locks but this means the "processing" status
	// update won't be visible to other workers and we lose visibility into
	// jobs that are actually under processing.
	//
	// RANDOM() is used to randomize record selection to minimize two workers
	// picking up the same job when there's high load.
	FindPendingJobSQL = `
		SELECT * FROM jobs
		WHERE status = 'pending'
			AND (scheduled_at IS NULL
				OR scheduled_at < CURRENT_TIMESTAMP)
		ORDER BY RANDOM()
		LIMIT 1;`

	UpdateJobStatusSQL = `
		UPDATE jobs
		SET status = $1,
			error = $2,
			updated_at = $3
		WHERE id = $4 AND status = $5
		RETURNING *;`
)

type JobStatus string

const (
	// Job is awaiting to be picked up by a worker
	PendingStatus JobStatus = "pending"
	// Job has been picked up by a worker and is currently running
	RunningStatus JobStatus = "running"
	// Job has been ran by a worker and failed
	FailedStatus JobStatus = "failed"
	// Job has been ran by a worker and completed
	CompletedStatus JobStatus = "completed"
)

type Job struct {
	ID      int64     `db:"id" json:"id"`
	Name    string    `db:"name" json:"name"`
	Status  JobStatus `db:"status" json:"status"`
	Payload string    `db:"payload" json:"payload"`
	Error   string    `db:"error" json:"error"`

	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	ScheduledAt time.Time `db:"scheduled_at" json:"scheduled_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ensureJobsTable creates the jobs table if absent. It is idempotent (all DDL is IF NOT
// EXISTS), so the primitives that may be the first jobs operation in a flow — scheduleJob,
// findPendingJobByName, takeOnePendingJob — call it lazily on entry. A job can then be
// scheduled or polled from any context, on a fresh database, without a prior setup step.
// (The mark* helpers only run after takeOnePendingJob, so the table already exists.)
func ensureJobsTable(ctx context.Context) error {
	return data.Exec(ctx, CreateJobsTableSQL)
}

func scheduleJob(ctx context.Context, name string, payload []byte, t time.Time) (*Job, error) {
	if err := ensureJobsTable(ctx); err != nil {
		return nil, err
	}
	if t.IsZero() {
		t = time.Now()
	}

	job := &Job{}
	err := data.Get(ctx, job, ScheduleJobSQL,
		name, PendingStatus, string(payload), t,
	)
	if err != nil {
		return nil, err
	} else {
		return job, nil
	}
}

// scheduleJobIfNotExists returns data's no-rows error when a pending job with the same
// name already exists.
func scheduleJobIfNotExists(ctx context.Context, name string, payload []byte, t time.Time) (*Job, error) {
	if err := ensureJobsTable(ctx); err != nil {
		return nil, err
	}
	if t.IsZero() {
		t = time.Now()
	}

	job := &Job{}
	err := data.Get(ctx, job, ScheduleJobIfNotExistsSQL,
		name, PendingStatus, string(payload), t,
	)
	if err != nil {
		return nil, err
	} else {
		return job, nil
	}
}

func takeOnePendingJob(ctx context.Context) (*Job, error) {
	if err := ensureJobsTable(ctx); err != nil {
		return nil, err
	}

	job := &Job{}
	err := data.Run(ctx, func(s data.Scope) error {
		if err := s.Get(job, FindPendingJobSQL); err != nil {
			return err
		} else if err := s.Exec(UpdateJobStatusSQL,
			RunningStatus, "", time.Now(),
			job.ID, job.Status,
		); err != nil {
			return err
		} else {
			return nil
		}
	})

	if data.IsNoRows(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		return job, nil
	}
}

func markJobAsFailed(ctx context.Context, jobId int64, reason string) error {
	return data.Exec(ctx, UpdateJobStatusSQL,
		FailedStatus, reason, time.Now(),
		jobId, RunningStatus)
}

func markJobAsCompleted(ctx context.Context, jobId int64) error {
	return data.Exec(ctx, UpdateJobStatusSQL,
		CompletedStatus, "", time.Now(),
		jobId, RunningStatus)
}
