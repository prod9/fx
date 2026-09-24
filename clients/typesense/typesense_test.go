package typesense

import (
	"context"
	"strings"
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxtest"
	"github.com/stretchr/testify/require"
	tsapi "github.com/typesense/typesense-go/v3/typesense/api"
)

func TestCollectionRoundTrip(t *testing.T) {
	cl := newTestClient(t)
	col := Collection{
		Name:   testCollectionName(t),
		Fields: []Field{{Name: "title", Type: StringType}},
	}

	require.NoError(t, cl.CreateCollection(t.Context(), col))
	require.NoError(t, cl.DestroyCollection(t.Context(), col))

	err := cl.DestroyCollection(t.Context(), col)
	require.True(t, IsNotFound(err), "destroying twice must report not found, got %v", err)
}

func TestCollectionSchemaReachesServer(t *testing.T) {
	cl := newTestClient(t)
	col := Collection{
		Name:        testCollectionName(t),
		DefaultSort: "rank",
		Fields: []Field{
			{Name: "title", Type: StringType, Infix: true, Locale: "th"},
			{Name: "rank", Type: Int32Type},
			{Name: "note", Type: StringType, Optional: true, NoIndex: true},
			{Name: "tags", Type: StringArrType, Optional: true},
			{Name: "place", Type: GeopointType, Optional: true},
			{Name: "meta", Type: ObjectType, Optional: true},
		},
	}
	createTestCollection(t, cl, col)

	got, err := cl.Raw().Collection(col.Name).Retrieve(t.Context())
	require.NoError(t, err)
	require.Equal(t, "rank", *got.DefaultSortingField)
	require.True(t, *got.EnableNestedFields, "object fields must enable nested fields")

	fields := map[string]tsapi.Field{}
	for _, f := range got.Fields {
		fields[f.Name] = f
	}
	require.Equal(t, "string", fields["title"].Type)
	require.True(t, *fields["title"].Infix)
	require.Equal(t, "th", *fields["title"].Locale)
	require.False(t, *fields["title"].Optional, "zero value keeps fields required")
	require.True(t, *fields["title"].Index, "zero value keeps fields indexed")
	require.Equal(t, "int32", fields["rank"].Type)
	require.True(t, *fields["note"].Optional)
	require.False(t, *fields["note"].Index)
	require.Equal(t, "string[]", fields["tags"].Type)
	require.Equal(t, "geopoint", fields["place"].Type)
	require.Equal(t, "object", fields["meta"].Type)
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

// createTestCollection creates col fresh, clearing any copy left by an aborted earlier
// run, and destroys it when the test ends.
func createTestCollection(t *testing.T, cl *Client, col Collection) {
	t.Helper()

	err := cl.DestroyCollection(t.Context(), col)
	if err != nil && !IsNotFound(err) {
		require.NoError(t, err)
	}
	require.NoError(t, cl.CreateCollection(t.Context(), col))

	t.Cleanup(func() {
		// t.Context() is already cancelled by the time cleanups run.
		require.NoError(t, cl.DestroyCollection(context.Background(), col))
	})
}

func testCollectionName(t *testing.T) string {
	return "fxtest_" + strings.ToLower(strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
}
