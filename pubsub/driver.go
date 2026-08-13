package pubsub

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"fx.prodigy9.co/config"
)

var (
	// URLConfig selects the backend driver: the URL's scheme picks a built-in driver,
	// the remainder is the driver's to interpret. Unset means the Postgres driver over
	// the ambient data context — exactly the pre-driver behavior.
	URLConfig = config.Str("PUBSUB_URL")

	ErrUnknownScheme = errors.New("pubsub: unsupported PUBSUB_URL scheme")
)

// driver is the internal backend seam, sitting exactly at the untyped floor: everything
// above it (Channel[T], the JSON codec, name and payload validation) is driver-agnostic,
// and the public API stays opaque about what runs underneath. Names and payloads arrive
// already validated. Implementations must make the initial Subscribe connect synchronous
// (error, never a live channel, on failure), reconnect silently, drop (never block) on a
// slow consumer, and close the returned channel only after the loop has fully stopped.
type driver interface {
	Publish(ctx context.Context, channel, payload string) error
	Subscribe(ctx context.Context, channel string) (<-chan string, context.CancelFunc, error)
}

type driverKey struct{}

// withDriver injects a driver directly, bypassing config resolution — the in-package
// test seam, letting unit tests run against an in-memory fake with no database.
func withDriver(ctx context.Context, d driver) context.Context {
	return context.WithValue(ctx, driverKey{}, d)
}

// resolveDriver picks the driver for one publish or subscribe from PUBSUB_URL on the ctx
// config source; unset falls back to Postgres over the ambient data context.
func resolveDriver(ctx context.Context) (driver, error) {
	if d, ok := ctx.Value(driverKey{}).(driver); ok {
		return d, nil
	}

	rawURL := ""
	if cfg := config.FromContext(ctx); cfg != nil {
		rawURL = config.Get(cfg, URLConfig)
	}
	if rawURL == "" {
		return postgresDriver{}, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("pubsub: invalid PUBSUB_URL: %w", err)
	}
	switch u.Scheme {
	case "postgres":
		return postgresDriver{}, nil
	case "redis", "rediss":
		return newRedisDriver(rawURL)
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownScheme, u.Scheme)
	}
}
