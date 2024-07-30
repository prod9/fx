package main

import (
	"context"
	"log"

	"fx.prodigy9.co/worker"
)

type Creator struct {
	Count int `json:"count"`
}

var _ worker.Interface = &Creator{}

func (r *Creator) Name() string { return "creator" }

func (r *Creator) Run(ctx context.Context) error {
	counters, err := CreateCounters(ctx, r.Count)
	if err != nil {
		return err
	}

	for _, c := range counters {
		log.Printf("created: %s\n", c.Name)
	}
	return nil
}
