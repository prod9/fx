// Package fxsentry reports errors and panics to Sentry through one process-wide client.
// The client is built from the config.Source of the first Activate call; contexts carry
// only nested scopes. Everything is a no-op while Sentry is off.
package fxsentry

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"fx.prodigy9.co/buildinfo"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxlog"
	"github.com/getsentry/sentry-go"
)

// flushTimeout matches sentry-go's own HTTP integration default.
const flushTimeout = 2 * time.Second

var (
	DSNConfig = config.Str("SENTRY_DSN").WithAlias("API_SENTRY_DSN")

	current      atomic.Pointer[sentry.Client]
	activateOnce sync.Once
)

type scopeKey struct{}

// Activate builds the client from cfg on the first call; later calls are ignored. It
// reports whether Sentry is on: off when SENTRY_DSN is unset or invalid.
func Activate(cfg *config.Source) bool {
	activateOnce.Do(func() {
		current.Store(newClient(cfg))
	})
	return current.Load() != nil
}

func newClient(cfg *config.Source) *sentry.Client {
	dsn := config.Get(cfg, DSNConfig)
	if dsn == "" {
		return nil
	}

	client, err := sentry.NewClient(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		fxlog.Errorf("fxsentry: %w", err)
		return nil
	}
	return client
}

// NewScope returns ctx carrying a new scope that inherits the scope already on ctx.
func NewScope(ctx context.Context) (context.Context, *sentry.Scope) {
	scope := scopeFrom(ctx)
	return context.WithValue(ctx, scopeKey{}, scope), scope
}

// Capture reports err with the context's scope.
func Capture(ctx context.Context, err error) {
	if client := current.Load(); client != nil {
		client.CaptureException(err, &sentry.EventHint{Context: ctx}, eventScope(ctx))
	}
}

// CapturePanic recovers a panic, reports it with the context's scope, flushes, and
// panics again. Defer it directly: `defer fxsentry.CapturePanic(ctx)`.
func CapturePanic(ctx context.Context) {
	recovered := recover()
	if recovered == nil {
		return
	}

	if client := current.Load(); client != nil {
		client.Recover(recovered, &sentry.EventHint{Context: ctx}, eventScope(ctx))
		client.Flush(flushTimeout)
	}
	panic(recovered)
}

// eventScope is the context's scope plus the build info: release and component name are
// set on every event, so sentry-go's own release fallbacks never apply.
func eventScope(ctx context.Context) *sentry.Scope {
	info := buildinfo.FromContext(ctx)
	scope := scopeFrom(ctx)
	scope.SetTag("component", info.Name)
	scope.AddEventProcessor(func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		event.Release = info.Version
		return event
	})
	return scope
}

// scopeFrom returns a copy of the context's scope, or a fresh one.
func scopeFrom(ctx context.Context) *sentry.Scope {
	if parent, ok := ctx.Value(scopeKey{}).(*sentry.Scope); ok {
		return parent.Clone()
	}
	return sentry.NewScope()
}
