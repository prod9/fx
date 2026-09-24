package typesense

import (
	"context"
	"strings"
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

// newTestClient talks to a real Typesense server. Tests skip when TYPESENSE_API_KEY
// is unset so `go test ./...` stays hermetic; set it in .env.local to run them.
func newTestClient(t *testing.T) *Client {
	t.Helper()

	cfg := fxtest.Configure()
	if config.Get(cfg, APIKeyConfig) == "" {
		t.Skip("typesense: TYPESENSE_API_KEY is unset")
	}
	return New(cfg)
}

// createTestCollection creates col fresh, clearing any copy left by an aborted earlier
// run, and deletes it when the test ends.
func createTestCollection(t *testing.T, cl *Client, col Collection) {
	t.Helper()

	err := cl.DeleteCollection(t.Context(), col.Name)
	if err != nil && !IsNotFound(err) {
		require.NoError(t, err)
	}
	require.NoError(t, cl.CreateCollection(t.Context(), col))

	t.Cleanup(func() {
		// t.Context() is already cancelled by the time cleanups run.
		require.NoError(t, cl.DeleteCollection(context.Background(), col.Name))
	})
}

func testCollectionName(t *testing.T) string {
	return "fxtest_" + strings.ToLower(strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
}
