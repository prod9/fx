package pubsub

import (
	"context"
	"encoding/json"

	"fx.prodigy9.co/fxlog"
)

// SubscribeRaw streams raw payloads on a channel by name — the untyped floor beneath
// Subscribe. It validates the name, resolves the driver (WithDriver → PUBSUB_URL →
// Postgres over the data context), and hands the subscription to it. The initial connect
// is synchronous: on failure it returns the error rather than a live channel, so the
// failure surfaces loud at the callsite instead of a silent never-delivering stream.
//
// A subscription owns its backend resources alone, so its whole lifetime is one context;
// it MUST be cancelled (defer the returned cancel), else the loop goroutine and whatever
// backend resource it holds leak for the life of the process.
func SubscribeRaw(ctx context.Context, name string) (<-chan string, context.CancelFunc, error) {
	if err := validateName(name); err != nil {
		return nil, nil, err
	}

	d, err := resolveDriver(ctx)
	if err != nil {
		return nil, nil, err
	}
	return d.Subscribe(ctx, name)
}

// Subscribe streams typed payloads on ch. A payload that fails to decode into T is
// logged and skipped — the consumer is built to re-derive from the table and never has
// to trust the payload.
func Subscribe[T any](ctx context.Context, ch Channel[T]) (<-chan T, context.CancelFunc, error) {
	raw, rawCancel, err := SubscribeRaw(ctx, ch.name)
	if err != nil {
		return nil, nil, err
	}

	// The decoder needs its own cancellation, not just the parent ctx: a caller that stops
	// reading and calls the returned cancel (without cancelling the ctx it passed in) would
	// otherwise leave the decoder blocked on a stalled send forever. cancel drives both the
	// driver's loop and the decoder.
	decodeCtx, decodeCancel := context.WithCancel(ctx)
	cancel := func() {
		rawCancel()
		decodeCancel()
	}

	out := make(chan T)
	go decode(decodeCtx, raw, out)
	return out, cancel, nil
}

// decode unmarshals each raw payload into T, logging and skipping any that fail. The send
// to out blocks until the consumer reads or ctx is cancelled; upstream, the driver has
// already dropped anything the stalled consumer couldn't keep up with.
func decode[T any](ctx context.Context, raw <-chan string, out chan<- T) {
	defer close(out)

	for payload := range raw {
		var value T
		if err := json.Unmarshal([]byte(payload), &value); err != nil {
			fxlog.Log("pubsub: dropping undecodable payload", fxlog.Any("err", err))
			continue
		}
		select {
		case out <- value:
		case <-ctx.Done():
			return
		}
	}
}
