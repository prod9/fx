package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/ctrlc"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/errutil"
	"fx.prodigy9.co/fxlog"
)

var (
	PollingIntervalConfig = config.DurationDef("WORKER_POLL", 1*time.Minute)

	ErrJobExists = errors.New("job already exists")
	ErrStop      = errors.New("stop requested")
)

type (
	Interface interface {
		Name() string
		Run(ctx context.Context) error
	}

	// Resetter marks the job as needing a Reset before a run.
	//
	// Job instances are reused across runs primarily to avoid requiring calls to the
	// `reflect` package, so some fields might contains stale data from previous runs. If
	// this is the case, the job should implement this interface and reset itself to a clean
	// state.
	Resetter interface {
		Reset()
	}

	Worker struct {
		sync.Mutex
		interval  time.Duration
		knownJobs map[string]Interface
		filters   []Filter
		cfg       *config.Source
		cancel    context.CancelCauseFunc
	}

	workerSignal int
)

const (
	signalIdled    workerSignal = iota
	signalWorkDone workerSignal = iota
	signalStop     workerSignal = iota
)

func New(cfg *config.Source, jobs ...Interface) *Worker {
	w := &Worker{
		interval: config.Get(cfg, PollingIntervalConfig),
		filters:  []Filter{SentryFilter, RecoverFilter, LogFilter},
		cfg:      cfg,
		cancel:   nil,
	}

	w.Register(jobs...)
	return w
}

func ScheduleNowIfNotExists(ctx context.Context, job Interface) (int64, error) {
	return ScheduleAtIfNotExists(ctx, job, time.Now())
}
func ScheduleInIfNotExists(ctx context.Context, job Interface, d time.Duration) (int64, error) {
	return ScheduleAtIfNotExists(ctx, job, time.Now().Add(d))
}
func ScheduleAtIfNotExists(ctx context.Context, job Interface, t time.Time) (int64, error) {
	payload, err := json.Marshal(job)
	if err != nil {
		return 0, err
	}

	scheduled, err := scheduleJobIfNotExists(ctx, job.Name(), payload, t)
	if data.IsNoRows(err) {
		return 0, ErrJobExists
	} else if err != nil {
		return 0, err
	}

	fxlog.Log("scheduling",
		fxlog.String("job", job.Name()),
		fxlog.Time("at", t),
	)
	return scheduled.ID, nil
}

func ScheduleNow(ctx context.Context, job Interface) (int64, error) {
	return ScheduleAt(ctx, job, time.Now())
}
func ScheduleIn(ctx context.Context, job Interface, d time.Duration) (int64, error) {
	return ScheduleAt(ctx, job, time.Now().Add(d))
}
func ScheduleAt(ctx context.Context, job Interface, t time.Time) (int64, error) {
	fxlog.Log("scheduling",
		fxlog.String("job", job.Name()),
		fxlog.Time("at", t),
	)

	if payload, err := json.Marshal(job); err != nil {
		return 0, err
	} else if job_, err := scheduleJob(ctx, job.Name(), payload, t); err != nil {
		return 0, err
	} else {
		return job_.ID, nil
	}
}

func (w *Worker) Register(jobs ...Interface) {
	w.Lock()
	defer w.Unlock()

	if w.knownJobs == nil {
		w.knownJobs = make(map[string]Interface)
	}
	for _, job := range jobs {
		w.knownJobs[job.Name()] = job
	}
}

func (w *Worker) Start() error {
	return w.StartContext(context.Background())
}

// StartContext runs the worker until stopped; job contexts descend from ctx.
func (w *Worker) StartContext(ctx context.Context) (err error) {
	defer errutil.Wrap("worker", &err)

	if w.cfg == nil {
		w.cfg = config.Configure()
	}

	db, err := data.Connect(w.cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithCancelCause(ctx)
	ctx = config.NewContext(ctx, w.cfg)
	ctx = data.NewContext(ctx, db)

	go func() {
		w.Lock()
		defer w.Unlock()

		if err = ensureJobsTable(ctx); err != nil {
			cancel(err)
			return
		}

		w.cancel = cancel
		go w.work(ctx)
	}()

	ctrlc.Do(w.Stop)

	fxlog.Log("worker started")
	<-ctx.Done()

	if ctx.Err() != nil {
		if err = context.Cause(ctx); err != nil {
			return err
		} else {
			return ctx.Err()
		}
	} else {
		return nil
	}
}

func (w *Worker) Stop() {
	w.Lock()
	defer w.Unlock()

	if w.cancel != nil {
		w.cancel(ErrStop)
	}
}

func (w *Worker) work(ctx context.Context) {
	for {
		// keep processing jobs, if there are jobs to process.
		// idle and poll only when there are no more jobs to process
		sig := w.workOnce(ctx)
		for sig == signalWorkDone {
			sig = w.workOnce(ctx)
		}

		switch sig {
		case signalStop:
			return
		case signalIdled:
			select {
			case <-ctx.Done():
				return
			case <-time.After(w.interval):
				continue
			}
		}
	}
}

func (w *Worker) workOnce(ctx context.Context) workerSignal {
	w.Lock()
	defer w.Unlock()

	job, err := takeOnePendingJob(ctx)
	if err != nil {
		w.cancel(err)
		return signalStop
	} else if job == nil {
		return signalIdled
	}

	if err = w.processJob(ctx, job); err != nil {
		err = markJobAsFailed(ctx, job.ID, err.Error())
	} else {
		err = markJobAsCompleted(ctx, job.ID)
	}
	if err != nil {
		w.cancel(err)
		return signalStop
	}
	return signalWorkDone
}

// processJob prepares the job, wraps it in the worker's filters, and runs it.
// Preparation is fixed: filters never see a job before its Reset and payload decode, so
// preparation failures are logged here rather than by LogFilter.
func (w *Worker) processJob(ctx context.Context, job *Job) error {
	instance, err := w.prepare(job)
	if err != nil {
		fxlog.Log("failed",
			fxlog.String("job", job.Name),
			fxlog.Int64("id", job.ID),
			fxlog.Any("error", err),
		)
		return err
	}

	for _, filter := range w.filters {
		if instance, err = filter(ctx, job.ID, instance); err != nil {
			return err
		}
	}

	if err = instance.Run(ctx); err != nil {
		return fmt.Errorf("run failed: %w", err)
	}
	return nil
}

func (w *Worker) prepare(job *Job) (Interface, error) {
	instance, ok := w.knownJobs[job.Name]
	if !ok {
		return nil, errors.New("unknown (or unregistered) job: " + job.Name)
	}

	if resetter, ok := instance.(Resetter); ok {
		resetter.Reset()
	}
	if err := json.Unmarshal([]byte(job.Payload), instance); err != nil {
		return nil, fmt.Errorf("malformed payload: %w", err)
	}
	return instance, nil
}
