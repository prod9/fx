package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxlog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var ErrNoDatabase = errors.New("pubsub: no database in context")

// SubscribeRaw opens a dedicated pooled connection, issues LISTEN, and streams raw
// payloads on the returned channel until the passed ctx is cancelled or the returned
// cancel is called — the untyped floor beneath Subscribe. The initial connect and LISTEN
// are synchronous: if either fails, it returns the error rather than a live channel, so
// the failure surfaces loud at the callsite instead of a silent never-delivering stream.
//
// The subscription owns its connection alone, so its whole lifetime is one context; a
// subscription MUST be cancelled (defer the returned cancel), else the loop goroutine and
// its held Postgres connection leak for the life of the process.
func SubscribeRaw(ctx context.Context, name string) (<-chan string, context.CancelFunc, error) {
	if err := validateName(name); err != nil {
		return nil, nil, err
	}

	db, ok := data.LookupFromContext(ctx)
	if !ok {
		return nil, nil, ErrNoDatabase
	}

	ctx, cancel := context.WithCancel(ctx)
	out := make(chan string)
	started := make(chan error, 1)

	go listen(ctx, db, name, out, started)

	if err := <-started; err != nil {
		cancel()
		return nil, nil, err
	}
	return out, cancel, nil
}

// Subscribe streams typed payloads on ch, backed by one dedicated LISTEN connection. A
// payload that fails to decode into T is logged and skipped — the consumer is built to
// re-derive from the table and never has to trust the payload.
func Subscribe[T any](ctx context.Context, ch Channel[T]) (<-chan T, context.CancelFunc, error) {
	raw, rawCancel, err := SubscribeRaw(ctx, ch.name)
	if err != nil {
		return nil, nil, err
	}

	// The decoder needs its own cancellation, not just the parent ctx: a caller that stops
	// reading and calls the returned cancel (without cancelling the ctx it passed in) would
	// otherwise leave the decoder blocked on a stalled send forever. cancel drives both the
	// listen loop and the decoder.
	decodeCtx, decodeCancel := context.WithCancel(ctx)
	cancel := func() {
		rawCancel()
		decodeCancel()
	}

	out := make(chan T)
	go decode(decodeCtx, raw, out)
	return out, cancel, nil
}

// listen drives one subscription across the whole lifetime of ctx. The first connect and
// LISTEN report their outcome once over started; after that, a dropped connection is
// reconnected silently and every other exit path is a clean shutdown.
func listen(ctx context.Context, db *sqlx.DB, name string, out chan<- string, started chan<- error) {
	defer close(out)

	first := true
	for {
		err := listenOnce(ctx, db, name, out, func() {
			if first {
				first = false
				started <- nil
			}
		})

		if first {
			if err == nil {
				err = ctx.Err()
			}
			started <- err
			return
		}
		if ctx.Err() != nil {
			return
		}

		fxlog.Log("pubsub: listen connection lost, reconnecting",
			fxlog.String("channel", name),
			fxlog.Any("err", err),
		)
		select {
		case <-ctx.Done():
			return
		case <-time.After(reconnectDelay):
		}
	}
}

// listenOnce holds one pooled connection for a single LISTEN session. The pgx handle is
// valid only inside conn.Raw, so LISTEN, the notification loop, and UNLISTEN all run in
// the one callback. onListen fires once LISTEN succeeds, before the loop blocks.
func listenOnce(ctx context.Context, db *sqlx.DB, name string, out chan<- string, onListen func()) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Raw(func(driverConn any) error {
		pgxConn := driverConn.(*stdlib.Conn).Conn()

		if _, err := pgxConn.Exec(ctx, "LISTEN "+pgx.Identifier{name}.Sanitize()); err != nil {
			return err
		}
		onListen()

		waitErr := pump(ctx, pgxConn, out)

		// Clear the registration only on a clean stop, when the connection is healthy
		// enough to return to the pool. On a dropped connection the pool discards it and
		// UNLISTEN would only error against a dead backend.
		if ctx.Err() != nil {
			if _, err := pgxConn.Exec(context.Background(), "UNLISTEN *"); err != nil {
				fxlog.Log("pubsub: UNLISTEN on release failed",
					fxlog.String("channel", name),
					fxlog.Any("err", err),
				)
			}
		}
		return waitErr
	})
}

// pump drains notifications to out with a non-blocking send — the one drop point. A
// consumer that can't keep up misses messages, the same at-most-once contract Postgres
// has when no one listens, and covered by the app's re-scan (§5). It never blocks the
// loop, so one stalled consumer can never back up the shared server-side notify queue.
func pump(ctx context.Context, conn *pgx.Conn, out chan<- string) error {
	for {
		notif, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		select {
		case out <- notif.Payload:
		default:
		}
	}
}

// decode unmarshals each raw payload into T, logging and skipping any that fail. The send
// to out blocks until the consumer reads or ctx is cancelled; upstream, pump has already
// dropped anything the stalled consumer couldn't keep up with.
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
