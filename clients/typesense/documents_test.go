package typesense

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testDoc struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Rank  int32  `json:"rank"`
}

func TestDocumentLifecycle(t *testing.T) {
	cl := newTestClient(t)
	col := createTestDocCollection(t, cl)
	ctx := t.Context()

	doc := testDoc{ID: "1", Title: "first", Rank: 1}
	require.NoError(t, cl.CreateDocument(ctx, col, doc))
	got, err := GetDocument[testDoc](ctx, cl, col, "1")
	require.NoError(t, err)
	require.Equal(t, doc, got)

	doc = testDoc{ID: "1", Title: "updated", Rank: 2}
	require.NoError(t, cl.UpdateDocument(ctx, col, "1", doc))
	got, err = GetDocument[testDoc](ctx, cl, col, "1")
	require.NoError(t, err)
	require.Equal(t, doc, got)

	require.NoError(t, cl.DestroyDocument(ctx, col, "1"))
	_, err = GetDocument[testDoc](ctx, cl, col, "1")
	require.True(t, IsNotFound(err), "destroyed document must be not found, got %v", err)
}

func TestIndexCreatesThenReplaces(t *testing.T) {
	cl := newTestClient(t)
	col := createTestDocCollection(t, cl)
	ctx := t.Context()

	require.NoError(t, cl.Index(ctx, col, testDoc{ID: "1", Title: "first", Rank: 1}))
	doc := testDoc{ID: "1", Title: "replaced", Rank: 5}
	require.NoError(t, cl.Index(ctx, col, doc))

	got, err := GetDocument[testDoc](ctx, cl, col, "1")
	require.NoError(t, err)
	require.Equal(t, doc, got)
}

func TestDocumentErrorsAreClassified(t *testing.T) {
	cl := newTestClient(t)
	col := createTestDocCollection(t, cl)
	ctx := t.Context()

	doc := testDoc{ID: "1", Title: "first", Rank: 1}
	require.NoError(t, cl.CreateDocument(ctx, col, doc))

	err := cl.CreateDocument(ctx, col, doc)
	require.True(t, IsConflict(err), "creating a taken id must be a conflict, got %v", err)

	err = cl.UpdateDocument(ctx, col, "missing", testDoc{ID: "missing", Title: "x"})
	require.True(t, IsNotFound(err), "updating a missing document must be not found, got %v", err)

	err = cl.DestroyDocument(ctx, col, "missing")
	require.True(t, IsNotFound(err), "destroying a missing document must be not found, got %v", err)
}

func createTestDocCollection(t *testing.T, cl *Client) Collection {
	t.Helper()

	col := Collection{
		Name: testCollectionName(t),
		Fields: []Field{
			{Name: "title", Type: StringType},
			{Name: "rank", Type: Int32Type},
		},
	}
	createTestCollection(t, cl, col)
	return col
}
