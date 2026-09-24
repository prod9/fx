package typesense

import (
	"strings"
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
)

func TestCollectionRoundTrip(t *testing.T) {
	cl := newTestClient(t)
	col := BuildCollection(testCollectionName(t)).
		Field("title", StringType, false, false).
		Build()

	require.NoError(t, cl.CreateCollection(t.Context(), col))
	require.NoError(t, cl.DestroyCollection(t.Context(), col))

	err := cl.DestroyCollection(t.Context(), col)
	require.True(t, IsNotFound(err), "destroying twice must report not found, got %v", err)
}

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

func testCollectionName(t *testing.T) string {
	return "fxtest_" + strings.ToLower(strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
}
