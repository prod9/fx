package typesense

import (
	"context"

	ts "github.com/typesense/typesense-go/v3/typesense"
	tsapi "github.com/typesense/typesense-go/v3/typesense/api"
)

// GetDocument fetches the document with the given id, decoded into T through its json
// tags.
func GetDocument[T any](ctx context.Context, cl *Client, col Collection, id string) (T, error) {
	return ts.GenericCollection[T](cl.ts, col.Name).Document(id).Retrieve(ctx)
}

// CreateDocument indexes doc as a new document; it fails with IsConflict when a
// document with the same id already exists.
func (cl *Client) CreateDocument(ctx context.Context, col Collection, doc any) error {
	_, err := cl.ts.Collection(col.Name).Documents().Create(ctx, doc, &tsapi.DocumentIndexParameters{})
	return err
}

// Index creates doc, or replaces the existing document with the same id.
func (cl *Client) Index(ctx context.Context, col Collection, doc any) error {
	_, err := cl.ts.Collection(col.Name).Documents().Upsert(ctx, doc, &tsapi.DocumentIndexParameters{})
	return err
}

// UpdateDocument writes doc over the existing document with the given id; it fails with
// IsNotFound when there is no such document.
func (cl *Client) UpdateDocument(ctx context.Context, col Collection, id string, doc any) error {
	_, err := cl.ts.Collection(col.Name).Document(id).Update(ctx, doc, &tsapi.DocumentIndexParameters{})
	return err
}

// DestroyDocument removes the document with the given id.
func (cl *Client) DestroyDocument(ctx context.Context, col Collection, id string) error {
	_, err := cl.ts.Collection(col.Name).Document(id).Delete(ctx)
	return err
}
