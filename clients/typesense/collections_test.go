package typesense

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectionSchemaRoundTrips(t *testing.T) {
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

	got, err := cl.GetCollection(t.Context(), col.Name)
	require.NoError(t, err)
	require.Equal(t, col, got)
}

func TestListCollectionsIncludesCreated(t *testing.T) {
	cl := newTestClient(t)
	col := Collection{Name: testCollectionName(t), Fields: []Field{{Name: "title", Type: StringType}}}
	createTestCollection(t, cl, col)

	cols, err := cl.ListCollections(t.Context())
	require.NoError(t, err)
	require.Contains(t, cols, col)
}

func TestCollectionErrorsAreClassified(t *testing.T) {
	cl := newTestClient(t)
	col := Collection{Name: testCollectionName(t), Fields: []Field{{Name: "title", Type: StringType}}}
	createTestCollection(t, cl, col)

	err := cl.CreateCollection(t.Context(), col)
	require.True(t, IsConflict(err), "creating a taken name must be a conflict, got %v", err)
	require.False(t, IsNotFound(err))

	_, err = cl.GetCollection(t.Context(), col.Name+"_missing")
	require.True(t, IsNotFound(err), "missing collection must be not found, got %v", err)
	require.False(t, IsConflict(err))
}
