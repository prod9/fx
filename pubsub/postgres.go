package pubsub

import (
	"context"
	"errors"
	"time"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxlog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var ErrNoDatabase = errors.New("pubsub: no database in context")

// reconnectDelay is the tight backoff between listen-connection reconnect attempts.
const reconnectDelay = 1 * time.Second

// postgresDriver is the default driver: LISTEN/NOTIFY over the *sqlx.DB already on the
// data context. It holds no state of its own — the database rides ctx, so it needs no
// URL and no process-wide registration. See docs/spec/pubsub-postgres.md.
type postgresDriver struct{}

// Publish runs through data.Exec, so it rides whatever transaction context it is called
// in: inside data.Run it joins the parent tx and fires on that commit; on its own it
// runs a BEGIN/pg_notify/COMMIT of its own. Prefer publishing out-of-band, after the
// business commit — coupling a NOTIFY into the business tx lets a full notify queue roll
// it back.
func (postgresDriver) Publish(ctx context.Context, channel, payload string) error {
	if _, ok := data.LookupFromContext(ctx); !ok {
		return ErrNoDatabase
	}

	return data.Exec(ctx, `SELECT pg_notify($1, $2)`, channel, payload)
}

// Subscribe opens a dedicated pooled connection, issues LISTEN, and streams raw payloads
// on the returned channel until the passed ctx is cancelled or the returned cancel is
// called. One subscription is one Postgres backend; the pool's DATABASE_MAX_OPEN is the
// budget it draws against.
func (postgresDriver) Subscribe(ctx context.Context, channel string) (<-chan string, context.CancelFunc, error) {
	db, ok := data.LookupFromContext(ctx)
	if !ok {
		return nil, nil, ErrNoDatabase
	}

	ctx, cancel := context.WithCancel(ctx)
	out := make(chan string)
	started := make(chan error, 1)

	go listen(ctx, db, channel, out, started)

	if err := <-started; err != nil {
		cancel()
		return nil, nil, err
	}
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
