package pubsub

import (
	"context"
	"net/url"
	"testing"
	"time"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxtest"

	"github.com/stretchr/testify/require"
)

// fakeDriver is an in-memory Driver: publishes record into published, subscriptions
// stream from the subs channel.
type fakeDriver struct {
	published map[string][]string
	subs      chan string
}

func newFakeDriver() *fakeDriver {
	return &fakeDriver{
		published: map[string][]string{},
		subs:      make(chan string, 16),
	}
}

func (f *fakeDriver) Publish(ctx context.Context, channel, payload string) error {
	f.published[channel] = append(f.published[channel], payload)
	return nil
}

func (f *fakeDriver) Subscribe(ctx context.Context, channel string) (<-chan string, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(ctx)
	out := make(chan string)
	go func() {
		defer close(out)
		for {
			select {
			case payload := <-f.subs:
				select {
				case out <- payload:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, cancel, nil
}

func TestWithDriverOverridesResolution(t *testing.T) {
	fake := newFakeDriver()
	ctx := WithDriver(context.Background(), fake)

	require.NoError(t, PublishRaw(ctx, "orders_changed", `{"id":1}`))
	require.Equal(t, []string{`{"id":1}`}, fake.published["orders_changed"])
}

func TestTypedRoundTripThroughDriver(t *testing.T) {
	type event struct {
		ID int64 `json:"id"`
	}
	ch := NewChannel[event]("driver_roundtrip")

	fake := newFakeDriver()
	ctx := WithDriver(context.Background(), fake)

	out, cancel, err := Subscribe(ctx, ch)
	require.NoError(t, err)
	defer cancel()

	require.NoError(t, Publish(ctx, ch, event{ID: 42}))
	require.Equal(t, []string{`{"id":42}`}, fake.published["driver_roundtrip"])

	fake.subs <- `{"id":42}`
	select {
	case got := <-out:
		require.Equal(t, event{ID: 42}, got)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for typed delivery")
	}
}

func TestUnknownSchemeIsResolutionError(t *testing.T) {
	cfg := fxtest.Configure()
	config.Set(cfg, URLConfig, "bogus://localhost")
	ctx := config.NewContext(context.Background(), cfg)

	err := PublishRaw(ctx, "orders_changed", "x")
	require.ErrorIs(t, err, ErrUnknownScheme)
}

func TestUnsetURLDefaultsToPostgres(t *testing.T) {
	// No driver, no PUBSUB_URL, no data context: the default path is the Postgres
	// driver over the ambient data context, so its absence is the error.
	err := PublishRaw(context.Background(), "orders_changed", "x")
	require.ErrorIs(t, err, ErrNoDatabase)
}

func TestSchemeDriverIsCachedPerURL(t *testing.T) {
	calls := 0
	RegisterScheme("cachedtest", func(cfg *config.Source, u *url.URL) (Driver, error) {
		calls++
		return newFakeDriver(), nil
	})

	cfg := fxtest.Configure()
	config.Set(cfg, URLConfig, "cachedtest://localhost")
	ctx := config.NewContext(context.Background(), cfg)

	require.NoError(t, PublishRaw(ctx, "orders_changed", "a"))
	require.NoError(t, PublishRaw(ctx, "orders_changed", "b"))
	require.Equal(t, 1, calls)
}
