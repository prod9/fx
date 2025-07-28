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
	// TODO: Might need to be careful with transactions here
	_, err := findPendingJobByName(ctx, job.Name())
	if data.IsNoRows(err) {
		return ScheduleAt(ctx, job, t)
	} else {
		return 0, ErrJobExists
	}
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
	} else if job, err := scheduleJob(ctx, job.Name(), payload, t); err != nil {
		return 0, err
	} else {
		return job.ID, nil
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

func (w *Worker) Start() (err error) {
	defer errutil.Wrap("worker", &err)

	if w.cfg == nil {
		w.cfg = config.Configure()
	}

	db, err := data.Connect(w.cfg)
	if err != nil {
		return err
	}

	var (
		ctx    context.Context
		cancel context.CancelCauseFunc
	)

	ctx, cancel = context.WithCancelCause(context.Background())
	ctx = config.NewContext(ctx, w.cfg)
	ctx = data.NewContext(ctx, db)

	go func() {
		w.Lock()
		defer w.Unlock()

		if err = createJobsTable(ctx); err != nil {
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
		// keep processing jobs, if there are jobs to process
		// idle and poll only when there are no more jobs to process
		sig := w.workOnce(ctx)
		for sig == signalWorkDone {
			sig = w.workOnce(ctx)
		}

		switch sig {
		case signalStop:
			return
		case signalIdled:
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(w.interval):
			continue
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

	fxlog.Log("running",
		fxlog.String("job", job.Name),
		fxlog.Int64("id", job.ID),
	)
	start := time.Now()

	// we got one "running" job to process
	if err := w.processJob(ctx, job); err != nil {
		fxlog.Log("failed",
			fxlog.String("job", job.Name),
			fxlog.Int64("id", job.ID),
			fxlog.Duration("duration", time.Since(start)),
			fxlog.Any("error", err),
		)
		if err := markJobAsFailed(ctx, job.ID, err.Error()); err != nil {
			w.cancel(err)
			return signalStop
		}

	} else {
		fxlog.Log("completed",
			fxlog.String("job", job.Name),
			fxlog.Int64("id", job.ID),
			fxlog.Duration("duration", time.Since(start)),
		)
		if err := markJobAsCompleted(ctx, job.ID); err != nil {
			w.cancel(err)
			return signalStop
		}
	}

	return signalWorkDone
}

// TODO: Add more speciailized errors for signaling retries/rerun
func (w *Worker) processJob(ctx context.Context, job *Job) error {
	var instance Interface

	if j, ok := w.knownJobs[job.Name]; !ok {
		return errors.New("unknown (or unregistered) job: " + job.Name)
	} else {
		instance = j
	}

	if resetter, ok := instance.(Resetter); ok {
		resetter.Reset()
	}
	if err := json.Unmarshal([]byte(job.Payload), instance); err != nil {
		return fmt.Errorf("malformed payload: %w", err)
	}

	// TODO: Enforce timeouts
	// TODO: Better to run the job in a separate transaction. So the job state is not
	// effected by the job code.
	if err := instance.Run(ctx); err != nil {
		return fmt.Errorf("run failed: %w", err)
	} else {
		return nil
	}
}
