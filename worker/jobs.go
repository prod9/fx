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
		);`

	FindPendingJobByNameSQL = `
		SELECT * FROM jobs
		WHERE name = $1
			AND status = 'pending'
		ORDER BY id DESC
		LIMIT 1;`
	ScheduleJobSQL = `
		INSERT INTO jobs (name, status, payload, scheduled_at)
		VALUES ($1, $2, $3, $4)
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

func createJobsTable(ctx context.Context) error {
	return data.Exec(ctx, CreateJobsTableSQL)
}

func findPendingJobByName(ctx context.Context, name string) (*Job, error) {
	job := &Job{}
	err := data.Get(ctx, job, FindPendingJobByNameSQL, name)
	if err != nil {
		return nil, err
	} else {
		return job, nil
	}
}

func scheduleJob(ctx context.Context, name string, payload []byte, t time.Time) (*Job, error) {
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

func takeOnePendingJob(ctx context.Context) (*Job, error) {
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
