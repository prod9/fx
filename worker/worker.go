package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/ctrlc"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/errutil"
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
	log.Println("scheduling", job.Name(), "at", t.Format(time.RFC3339))

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

	log.Println("worker started")
	<-ctx.Done()
	return ctx.Err()
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

	log.Printf("running %s #%d", job.Name, job.ID)
	start := time.Now()

	// we got one "running" job to process
	if err := w.processJob(ctx, job); err != nil {
		log.Printf("failed %s #%d in %s: %s\n",
			job.Name, job.ID,
			time.Now().Sub(start).String(), err.Error(),
		)
		if err := markJobAsFailed(ctx, job.ID, err.Error()); err != nil {
			w.cancel(err)
			return signalStop
		}

	} else {
		log.Printf("completed %s #%d in %s",
			job.Name, job.ID,
			time.Now().Sub(start).String(),
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
