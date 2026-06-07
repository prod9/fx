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

## Job Interface

Jobs implement the `worker.Interface`:

```go
type Interface interface {
	Name() string           // unique job name, used as DB key
	Run(ctx context.Context) error
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

## Configuration

* `WORKER_POLL` — Polling interval (default: `1m`).
