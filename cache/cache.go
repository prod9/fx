package cache

import (
	"context"
	"time"
)

type (
	Initializer[T any] func() (T, time.Duration, error)

	Interface[T any] interface {
		Get(ctx context.Context, initer Initializer[T]) (T, error)
		Invalidate(ctx context.Context) error
	}
)
