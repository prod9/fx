package fxsentry

import (
	"context"
	"errors"
	"testing"

	"fx.prodigy9.co/buildinfo"
	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func TestCapture_Off(t *testing.T) {
	current.Store(nil)
	require.NotPanics(t, func() { Capture(context.Background(), errBoom) })
}

func TestCapture_BuildInfoFromContext(t *testing.T) {
	transport := useMockClient(t)
	ctx := buildinfo.NewContext(context.Background(), buildinfo.Info{Name: "api", Version: "v1.2.3"})

	Capture(ctx, errBoom)

	events := transport.Events()
	require.Len(t, events, 1)
	require.Equal(t, "v1.2.3", events[0].Release)
	require.Equal(t, "api", events[0].Tags["component"])
}

func TestCapture_NestedScopesInherit(t *testing.T) {
	transport := useMockClient(t)
	outer, outerScope := NewScope(context.Background())
	outerScope.SetTag("outer", "yes")
	inner, innerScope := NewScope(outer)
	innerScope.SetTag("inner", "yes")

	Capture(inner, errBoom)
	Capture(outer, errBoom)

	events := transport.Events()
	require.Len(t, events, 2)
	require.Equal(t, "yes", events[0].Tags["outer"])
	require.Equal(t, "yes", events[0].Tags["inner"])
	require.Equal(t, "yes", events[1].Tags["outer"])
	require.NotContains(t, events[1].Tags, "inner")
}

func TestCapturePanic_ReportsAndRepanics(t *testing.T) {
	transport := useMockClient(t)

	require.PanicsWithValue(t, "kaboom", func() {
		defer CapturePanic(context.Background())
		panic("kaboom")
	})
	require.Len(t, transport.Events(), 1)
}

func TestCapturePanic_NoPanic(t *testing.T) {
	transport := useMockClient(t)

	func() {
		defer CapturePanic(context.Background())
	}()
	require.Empty(t, transport.Events())
}

func useMockClient(t *testing.T) *sentry.MockTransport {
	t.Helper()

	transport := &sentry.MockTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://key@sentry.example/1",
		Transport: transport,
	})
	require.NoError(t, err)

	current.Store(client)
	t.Cleanup(func() { current.Store(nil) })
	return transport
}
