package worker

import "context"

type Stub struct {
	JobName string
	RunFunc func(ctx context.Context) error
}

var _ Interface = Stub{}

func (s Stub) Name() string                  { return s.JobName }
func (s Stub) Run(ctx context.Context) error { return s.RunFunc(ctx) }
