package worker

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type TestJob struct {
	Arg          string    `json:"arg"`
	RegisteredAt time.Time `json:"registered_at"`
	Error        string    `json:"error"`
}

var _ Interface = &TestJob{}
var _ Resetter = &TestJob{}

func (r *TestJob) Name() string { return "test-job" }
func (r *TestJob) Reset() {
	r.Arg = ""
	r.RegisteredAt = time.Time{}
	r.Error = ""
}

func (r *TestJob) Run(ctx context.Context) error {
	defer fmt.Println("TestRunner finished")
	fmt.Printf(
		"TestRunner with arg `%s` registered at `%s`\n",
		r.Arg,
		r.RegisteredAt,
	)

	select {
	case <-ctx.Done():
		fmt.Printf("context cancelled: %s\n", ctx.Err())
	case <-time.After(500 * time.Millisecond):
		// continue
	}

	if r.Error != "" {
		return fmt.Errorf("mock error: %w", errors.New(r.Error))
	} else {
		return nil
	}
}
