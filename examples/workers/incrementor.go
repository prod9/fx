package main

import (
	"context"
	"time"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/worker"
)

type Incrementer struct {
	Counter     string `json:"name"`
	IncrementBy int    `json:"increment_by"`
}

var _ worker.Interface = &Incrementer{}

func (r *Incrementer) Name() string { return "incrementer" }

func (r *Incrementer) Run(ctx context.Context) (err error) {
	scope, cancel, err := data.NewScopeErr(ctx, &err)
	if err != nil {
		return err
	}
	defer cancel()

	counter, err := GetCounterByName(scope.Context(), r.Counter)
	if err != nil {
		return err
	}

	counter.Count += int64(r.IncrementBy)
	if err = UpdateCounter(scope.Context(), counter); err != nil {
		return err
	}

	fxlog.Log("updated "+counter.Name, fxlog.Int64("to", counter.Count))
	worker.ScheduleIn(ctx, r, 1*time.Second)
	return nil
}
