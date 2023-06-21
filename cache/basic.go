package cache

import (
	"context"
	"sync"
	"time"
)

type basic[T any] struct {
	mutex sync.RWMutex

	data    T
	expires time.Time
}

func Basic[T any]() Interface[T] {
	return &basic[T]{}
}

func (c *basic[T]) Get(_ context.Context, initer Initializer[T]) (T, error) {
	if data, ok := c.get(); ok {
		return data, nil
	} else {
		return c.initialize(initer)
	}
}

func (c *basic[T]) Invalidate(_ context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.expires = time.Now()
	return nil
}

func (c *basic[T]) get() (result T, ok bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if time.Now().Before(c.expires) {
		result, ok = c.data, true
	} else {
		ok = false
	}
	return
}

func (c *basic[T]) initialize(initer Initializer[T]) (result T, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var age time.Duration
	if result, age, err = initer(); err != nil {
		return
	}

	c.data = result
	c.expires = time.Now().Add(age)
	return
}
