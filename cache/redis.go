package cache

import (
	"context"
	"encoding/gob"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"fx.prodigy9.co/config"
	goredis "github.com/redis/go-redis/v9"
)

var RedisURLConfig = config.Str("REDIS_URL")

type redis[T any] struct {
	mutex sync.RWMutex // lock *redis initialization

	key   string
	cfg   *config.Source
	redis *goredis.Client
}

var errNotConnected = errors.New("redis is not connected")

const DefaultKey = "cache"

// Redis returns a redis-backed cache. The cache is lazily initialized. The first request
// to the cache primes it. Subsequent call gets the cached value
//
// If there is a problem during one or more connections (GET returns error), the cache
// will assume that redis is down or it has become unreachable and will automatically
// disconnect in the background so that other reads will continue to work through the cache.
//
// It will try to re-connect again on next request if it is not already connected.
func Redis[T any](cfg *config.Source, key string) Interface[T] {
	key = strings.TrimSpace(key)
	if len(key) == 0 {
		key = DefaultKey
	}

	return &redis[T]{
		key:   key,
		cfg:   cfg,
		redis: nil,
	}
}

func (r *redis[T]) Get(ctx context.Context, initer Initializer[T]) (result T, err error) {
	result, err = r.get(ctx)
	if errors.Is(err, goredis.Nil) {
		if result, err = r.initialize(ctx, initer); err != nil {
			return r.fallback(ctx, initer, err)
		}

	} else if errors.Is(err, errNotConnected) {
		if err = r.connect(ctx); err != nil {
			return r.fallback(ctx, initer, err)
		}

		result, err = r.get(ctx)
		if errors.Is(err, goredis.Nil) {
			if result, err = r.initialize(ctx, initer); err != nil {
				return r.fallback(ctx, initer, err)
			} // else no error
		} else if err != nil {
			return r.fallback(ctx, initer, err)
		} // else no error
	} // else no error
	return
}

func (r *redis[T]) fallback(ctx context.Context, initer Initializer[T], err error) (result T, outerr error) {
	if err == nil {
		result, _, outerr = initer()
		return
	}

	// client seems to be faulty, try to disconnect in the background so we re-connect
	// again on the next request
	go r.disconnect(context.Background())
	log.Println("redis cache:", err)
	result, _, outerr = initer()
	return
}

func (r *redis[T]) Invalidate(ctx context.Context) (err error) {
	err = r.invalidate(ctx)
	if errors.Is(err, errNotConnected) {
		if err = r.connect(ctx); err != nil {
			return
		} // else no error
		err = r.invalidate(ctx)
	} // else no error
	return
}

func (r *redis[T]) connect(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.redis != nil { // already connected
		return nil
	}

	opts, err := goredis.ParseURL(config.Get(r.cfg, RedisURLConfig))
	if err != nil {
		return err
	}

	client := goredis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return err
	} else {
		r.redis = client
		return nil
	}
}

func (r *redis[T]) disconnect(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.redis == nil {
		return nil
	} else {
		return r.redis.Close()
	}
}

func (r *redis[T]) get(ctx context.Context) (result T, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.redis == nil {
		err = errNotConnected
		return
	}

	var str string
	if str, err = r.redis.Get(ctx, r.key).Result(); err != nil {
		return
	}

	err = gob.NewDecoder(strings.NewReader(str)).Decode(&result)
	return
}

func (r *redis[T]) initialize(ctx context.Context, initer Initializer[T]) (result T, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.redis == nil {
		err = errNotConnected
		return
	}

	var age time.Duration
	if result, age, err = initer(); err != nil {
		return
	}

	str := &strings.Builder{}
	if err = gob.NewEncoder(str).Encode(&result); err != nil {
		return
	}

	err = r.redis.Set(ctx, r.key, str.String(), age).Err()
	return
}

func (r *redis[T]) invalidate(ctx context.Context) (err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.redis == nil {
		err = errNotConnected
		return
	}

	err = r.redis.Del(ctx, r.key).Err()
	if errors.Is(err, goredis.Nil) { // nothing to invalidate, not an error
		err = nil
	}
	return
}
