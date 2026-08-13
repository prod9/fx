# Background Workers

**Status:** accepted

The `worker` package provides a PostgreSQL-backed background job system.

## Setup

Register job types and start the worker:

```go
worker := worker.New(cfg, &SendEmailJob{}, &CleanupJob{})
worker.Start() // blocks, polling for jobs
worker.Stop()  // graceful shutdown
```

`Start` opens its own database pool and closes it when it returns, so start/stop
cycles do not accumulate connections.

## Job Interface

Jobs implement the `worker.Interface`:

```go
type Interface interface {
	Name() string           // unique job name, used as DB key
	Run(ctx context.Context) error
}
```

### Timeouts

The worker enforces no deadline on `Run` — timing out is the job's own decision to
make. A job doing potentially long-running work should think about its own timeout
(e.g. `context.WithTimeout` inside `Run`); a job that never returns stalls the
worker's processing loop.

### Retries

The worker has no retry mechanism — any non-nil error from `Run` marks the job
failed, permanently. A job that detects a retryable condition reschedules itself
before returning, the same way recurring jobs loop:

```go
func (j *SyncJob) Run(ctx context.Context) error {
	if err := sync(ctx); isRetryable(err) {
		if _, schedErr := worker.ScheduleIn(ctx, j, 5*time.Minute); schedErr != nil {
			return errors.Join(err, schedErr)
		}
		return err // this run still records as failed; the retry is a new job
	} else {
		return err
	}
}
```

## Scheduling

```go
worker.ScheduleNow(ctx, &SendEmailJob{To: "user@example.com"})
worker.ScheduleIn(ctx, &CleanupJob{}, 30*time.Minute)
worker.ScheduleAt(ctx, &ReportJob{}, tomorrow)

// schedule only if a pending job with the same name doesn't already exist
worker.ScheduleNowIfNotExists(ctx, &DailyDigestJob{})
```

The `IfNotExists` variants check-and-insert in a single SQL statement, so concurrent
schedulers of the same name cannot both win through a separate lookup. A sliver of a
race remains under READ COMMITTED (statements executing at the same instant share a
snapshot); a unique index would close it but would also forbid deliberately
scheduling duplicate pending jobs via `ScheduleAt`, so the sliver is accepted.

## Job state

Job tracking (claiming a job, marking it completed or failed) always runs on its own
transactions, separate from anything the job body does — a job's own rollback can
never roll back its status. Claiming is a CAS on the status column
(`UPDATE … WHERE id = $n AND status = 'pending'`) that commits immediately, so a
claim is visible to other workers at once and no transaction stays pinned open for
the job's runtime. `FOR UPDATE` locks were rejected for exactly that visibility
reason: a locked-row claim would hide which jobs are actually being processed. Jobs
wanting transactional work use `data.Run` inside `Run` themselves.

## Configuration

* `WORKER_POLL` — Polling interval (default: `1m`).
