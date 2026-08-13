package pubsub

import (
	"context"
	"testing"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxtest"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

const testRedisURL = "redis://localhost:6379/9"

// redisContext builds a config context selecting the redis driver, skipping the test
// when no Redis is reachable at testRedisURL.
func redisContext(t *testing.T) context.Context {
	opts, err := goredis.ParseURL(testRedisURL)
	require.NoError(t, err)
	probe := goredis.NewClient(opts)
	defer probe.Close()
	if err := probe.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis unreachable at %s: %s", testRedisURL, err)
	}

	cfg := fxtest.Configure()
	config.Set(cfg, URLConfig, testRedisURL)
	return config.NewContext(context.Background(), cfg)
}

func TestRedisSchemeResolves(t *testing.T) {
	// Resolution alone must pick the redis driver; a bogus host only fails later, on
	// the actual publish, with a network error — never ErrUnknownScheme.
	cfg := fxtest.Configure()
	config.Set(cfg, URLConfig, "redis://localhost:1/0")
	ctx := config.NewContext(context.Background(), cfg)

	err := PublishRaw(ctx, "orders_changed", "x")
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrUnknownScheme)
	require.NotErrorIs(t, err, ErrNoDatabase)
}

func TestRedisRoundTrip(t *testing.T) {
	type event struct {
		ID int64 `json:"id"`
	}
	ch := NewChannel[event]("redis_roundtrip")
	ctx := redisContext(t)

	out, cancel, err := Subscribe(ctx, ch)
	require.NoError(t, err)
	defer cancel()

	require.NoError(t, Publish(ctx, ch, event{ID: 7}))

	select {
	case got := <-out:
		require.Equal(t, event{ID: 7}, got)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for redis delivery")
	}
}

func TestRedisSubscribeCancelClosesChannel(t *testing.T) {
	ctx := redisContext(t)

	out, cancel, err := SubscribeRaw(ctx, "redis_cancel")
	require.NoError(t, err)

	cancel()
	select {
	case _, open := <-out:
		require.False(t, open, "channel must close after cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}
