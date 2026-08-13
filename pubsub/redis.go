package pubsub

import (
	"context"
	"errors"

	"fx.prodigy9.co/clients/redis"
	"fx.prodigy9.co/fxlog"

	goredis "github.com/redis/go-redis/v9"
)

// redisDriver is the driver for PUBSUB_URL=redis://…, built on native Redis PUB/SUB —
// the identical at-most-once shape, with connections far cheaper than Postgres backends,
// so it is the sanctioned driver for websocket-scale subscription counts. It holds the
// shared per-URL client from clients/redis; go-redis owns all connection mechanics,
// including the dedicated connection each subscription rides and its reconnect and
// resubscribe. See docs/spec/pubsub-redis.md.
type redisDriver struct {
	client *goredis.Client
}

func newRedisDriver(url string) (redisDriver, error) {
	client, err := redis.Client(url)
	if err != nil {
		return redisDriver{}, err
	}
	return redisDriver{client: client}, nil
}

// Publish is a plain PUBLISH — never transactional: it fires immediately even if a
// surrounding data.Run later rolls back. Apps following the out-of-band rule (publish
// after the commit) are unaffected.
func (d redisDriver) Publish(ctx context.Context, channel, payload string) error {
	return d.client.Publish(ctx, channel, payload).Err()
}

// Subscribe opens one go-redis subscription and adapts it to the package contract: the
// initial connect is synchronous (Receive surfaces the error before a channel is
// returned), the drain drops non-blocking to a slow consumer, and the out channel closes
// only after the loop has fully stopped.
func (d redisDriver) Subscribe(ctx context.Context, channel string) (<-chan string, context.CancelFunc, error) {
	sub := d.client.Subscribe(ctx, channel)
	if _, err := sub.Receive(ctx); err != nil {
		return nil, nil, errors.Join(err, sub.Close())
	}

	ctx, cancel := context.WithCancel(ctx)
	out := make(chan string)

	go func() {
		defer close(out)
		defer func() {
			if err := sub.Close(); err != nil {
				fxlog.Log("pubsub: redis subscription close failed",
					fxlog.String("channel", channel),
					fxlog.Any("err", err),
				)
			}
		}()

		in := sub.Channel()
		for {
			select {
			case msg, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- msg.Payload:
				default:
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, cancel, nil
}
