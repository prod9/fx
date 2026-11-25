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
cycles do not accumulate connections. `StartContext(ctx)` does the same with job
contexts descending from `ctx`; the `worker` command passes the app's context, which
carries `buildinfo` (see [buildinfo.md](buildinfo.md)).

## Filters

The worker runs one job at a time; scaling is by process. Each job goes through three
stages:

1. **Preparation, fixed:** resolve the job by name, `Reset` it, decode its payload. A
   failure here is logged as `failed` (job name, ID, error) by the worker itself, since
   no filter has run yet.
2. **Filters:** each `worker.Filter` (`func(ctx, id int64, job Interface) (Interface,
   error)`) takes the job's ID and the prepared job and returns the job, usually
   wrapped in a `worker.Stub{JobName, RunFunc}` whose `Run` defers around the inner
   `Run`. Filters apply in list order, each wrapping the result of the one before, so the
   last is outermost. A filter that returns an error fails the job without running it.
3. **Settle, fixed:** mark the job failed or completed from the run's result.

`worker.New` installs `SentryFilter`, `RecoverFilter` and `LogFilter`, in that order;
`SetFilters(...)` replaces the list.

* `SentryFilter` — when `SENTRY_DSN` is set, scopes the run and reports its panic or
  returned error. It sits before `RecoverFilter` so it sees the panic first.
* `RecoverFilter` — turns a panic into the run's error, so the job is marked failed and
  the worker keeps running.
* `LogFilter` — logs `running`, then `completed` or `failed` with the job's name, ID
  and duration.

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

## Cleanup

Every run of a recurring job leaves a finished row behind, so the `jobs` table grows
without bound. `worker.CleanupJob` keeps it bounded by deleting rows that need no further
processing. Register it and seed it once, like any self-rescheduling job:

```go
w := worker.New(cfg, &SendEmailJob{}, worker.CleanupJob)
worker.ScheduleNowIfNotExists(ctx, worker.CleanupJob)
```

Each run does two things:

1. A `running` row whose `updated_at` is older than `WORKER_STALE_DURATION` is marked
   `failed` with error `abandoned`. FX has no heartbeat, so a worker that died mid-job
   would otherwise leave the row `running` forever. The default of one hour treats any
   job running longer as a fault worth an operator's look; an app whose jobs
   legitimately run longer raises it. This is the only cleanup setting: it is the one
   value that depends on the app's jobs.
2. `completed` and `failed` rows older than a fixed retention of one day are deleted,
   except the latest `failed` row of each job name (by `id`), which is kept until a newer
   failure of that job replaces it. Any retention window would hide a job that fails
   unnoticed for longer than the window; keeping one failure per name keeps that evidence
   while the table stays bounded by the number of job names, not the number of runs.
   `pending` rows are never touched.

The deletion is a single `DELETE`; there is no batching knob, since Postgres handles the
volume. The job reschedules itself every two stale durations (two hours by default):
abandonment is the exception a sound worker setup rarely produces, so a crashed job may
show `running` for up to three stale durations. The job assumes a setup that runs the
worker regularly and makes no attempt to catch up on missed runs.

Alerting on failures is the job of `SentryFilter` and `LogFilter`, not of the `jobs`
table. A recurring job that keeps failing keeps running; FX has no circuit breaker yet.

`<app> worker cleanup` runs the same cleanup synchronously and exits, without a worker
process or scheduling anything.

## Configuration

* `WORKER_POLL` — Polling interval (default: `1m`).
* `WORKER_STALE_DURATION` — How long a job may stay `running` before cleanup marks it
  abandoned; cleanup runs every two of these (default: `1h`).
