package main

import (
	"context"
	"log"
	"time"

	"fx.prodigy9.co/worker"
)

type Reporter struct{}

var _ worker.Interface = &Reporter{}

func (r *Reporter) Name() string { return "reporter" }
func (r *Reporter) Run(ctx context.Context) error {
	counters, err := GetAllCounters(ctx)
	if err != nil {
		return err
	}

	log.Printf("reporting %d counters", len(counters))
	for _, c := range counters {
		log.Printf("%s: %d", c.Name, c.Count)
	}

	_, err = worker.ScheduleIn(ctx, r, 3*time.Second)
	return err
}
