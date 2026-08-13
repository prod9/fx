package pubsub

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"fx.prodigy9.co/config"
)

var (
	// URLConfig selects the backend driver: the URL's scheme picks the driver from the
	// registry, the remainder is the driver's to interpret. Unset means the Postgres
	// driver over the ambient data context — exactly the pre-driver behavior.
	URLConfig = config.Str("PUBSUB_URL")

	ErrUnknownScheme = errors.New("pubsub: no driver registered for scheme")

	registryMutex sync.Mutex
	factories     = map[string]func(cfg *config.Source, u *url.URL) (Driver, error){}
	drivers       = map[string]Driver{} // resolved drivers, cached per URL
)

// Driver is the backend seam, sitting exactly at the untyped floor: everything above it
// (Channel[T], the JSON codec, name and payload validation) is driver-agnostic. Names
// and payloads arrive already validated.
type Driver interface {
	// Publish sends payload on the named channel, best-effort, at-most-once.
	Publish(ctx context.Context, channel, payload string) error

	// Subscribe streams payloads on the named channel until ctx is cancelled or the
	// returned cancel is called. The initial connect is synchronous: on failure it
	// returns the error, never a live channel. Implementations must reconnect silently,
	// drop (never block) on a slow consumer, and close the returned channel only after
	// the loop has fully stopped.
	Subscribe(ctx context.Context, channel string) (<-chan string, context.CancelFunc, error)
}

type driverKey struct{}

// WithDriver injects a driver directly, bypassing config resolution — also the test
// seam: an in-memory Driver fake needs no database and no config.
func WithDriver(ctx context.Context, d Driver) context.Context {
	return context.WithValue(ctx, driverKey{}, d)
}

// RegisterScheme maps a PUBSUB_URL scheme to a driver factory. postgres and redis are
// pre-registered; an app plugs in its own backend by registering a scheme at init.
func RegisterScheme(scheme string, factory func(cfg *config.Source, u *url.URL) (Driver, error)) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	factories[scheme] = factory
}

// resolveDriver picks the driver for one publish or subscribe: a WithDriver injection
// wins, then the PUBSUB_URL scheme via the registry (constructed lazily, cached per
// URL), and unset falls back to Postgres over the ambient data context — exactly the
// pre-driver behavior.
func resolveDriver(ctx context.Context) (Driver, error) {
	if d, ok := ctx.Value(driverKey{}).(Driver); ok {
		return d, nil
	}

	rawURL := ""
	if cfg := config.FromContext(ctx); cfg != nil {
		rawURL = config.Get(cfg, URLConfig)
	}
	if rawURL == "" {
		return postgresDriver{}, nil
	}

	registryMutex.Lock()
	defer registryMutex.Unlock()
	if d, ok := drivers[rawURL]; ok {
		return d, nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("pubsub: invalid PUBSUB_URL: %w", err)
	}
	factory, ok := factories[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownScheme, u.Scheme)
	}

	d, err := factory(config.FromContext(ctx), u)
	if err != nil {
		return nil, err
	}
	drivers[rawURL] = d
	return d, nil
}
