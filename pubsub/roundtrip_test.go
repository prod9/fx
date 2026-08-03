package pubsub

import (
	"context"
	"testing"
	"time"

	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

type ball struct {
	Count int
}

func TestPublishSubscribeRawRoundTrip(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	ch, cancel, err := SubscribeRaw(ctx, "raw_roundtrip")
	require.NoError(t, err)
	defer cancel()

	got := recvRaw(t, ctx, "raw_roundtrip", "hello", ch)
	require.Equal(t, "hello", got)
}

func TestPublishSubscribeTypedRoundTrip(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	channel := NewChannel[ball]("typed_roundtrip")
	ch, cancel, err := Subscribe(ctx, channel)
	require.NoError(t, err)
	defer cancel()

	got := recvTyped(t, ctx, channel, ball{Count: 7}, ch)
	require.Equal(t, ball{Count: 7}, got)
}

// A payload that cannot decode into T is logged and skipped, not fatal: the stream stays
// live and a subsequent well-formed payload still arrives.
func TestDecodeFailureKeepsStreamLive(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	channel := NewChannel[ball]("decode_skip")
	ch, cancel, err := Subscribe(ctx, channel)
	require.NoError(t, err)
	defer cancel()

	require.NoError(t, PublishRaw(ctx, "decode_skip", "not-json"))

	got := recvTyped(t, ctx, channel, ball{Count: 42}, ch)
	require.Equal(t, ball{Count: 42}, got)
}

// Cancelling a typed subscription must close its channel, tearing down both the listen
// loop and the decoder goroutine.
func TestTypedSubscribeCancelClosesChannel(t *testing.T) {
	ctx := fxtest.ConnectTestDatabase(t)

	ch, cancel, err := Subscribe(ctx, NewChannel[ball]("cancel_close"))
	require.NoError(t, err)

	cancel()

	select {
	case _, ok := <-ch:
		require.False(t, ok, "channel must be closed after cancel")
	case <-time.After(2 * time.Second):
		t.Fatal("pubsub: typed channel not closed after cancel")
	}
}

// recvRaw publishes on a tick until one payload is delivered, absorbing the best-effort
// non-blocking drop: a publish that lands before the receiver is parked on the channel is
// simply dropped, and the next tick re-sends. Fails the test if nothing arrives in time.
func recvRaw(t *testing.T, ctx context.Context, name string, payload string, ch <-chan string) string {
	t.Helper()

	received := make(chan string, 1)
	go func() {
		select {
		case p := <-ch:
			received <- p
		case <-ctx.Done():
		}
	}()

	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	deadline := time.After(2 * time.Second)
	for {
		require.NoError(t, PublishRaw(ctx, name, payload))
		select {
		case p := <-received:
			return p
		case <-tick.C:
		case <-deadline:
			t.Fatal("pubsub: no delivery within deadline")
			return ""
		}
	}
}

func recvTyped[T any](t *testing.T, ctx context.Context, ch Channel[T], payload T, out <-chan T) T {
	t.Helper()

	received := make(chan T, 1)
	go func() {
		select {
		case v := <-out:
			received <- v
		case <-ctx.Done():
		}
	}()

	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	deadline := time.After(2 * time.Second)
	for {
		require.NoError(t, Publish(ctx, ch, payload))
		select {
		case v := <-received:
			return v
		case <-tick.C:
		case <-deadline:
			t.Fatal("pubsub: no typed delivery within deadline")
			var zero T
			return zero
		}
	}
}
