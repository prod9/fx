package typesense

import (
	"context"

	tsapi "github.com/typesense/typesense-go/v3/typesense/api"
)

func (cl *Client) CreateCollection(ctx context.Context, col Collection) error {
	_, err := cl.ts.Collections().Create(ctx, col.schema())
	return err
}

func (cl *Client) GetCollection(ctx context.Context, name string) (Collection, error) {
	resp, err := cl.ts.Collection(name).Retrieve(ctx)
	if err != nil {
		return Collection{}, err
	}
	return collectionFromResponse(resp)
}

func (cl *Client) ListCollections(ctx context.Context) ([]Collection, error) {
	resps, err := cl.ts.Collections().Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	cols := make([]Collection, len(resps))
	for i, resp := range resps {
		if cols[i], err = collectionFromResponse(resp); err != nil {
			return nil, err
		}
	}
	return cols, nil
}

// AddFields adds fields to an existing collection. Typesense only alters a schema by
// adding and dropping fields; to change a field, drop it and add it back.
func (cl *Client) AddFields(ctx context.Context, name string, fields ...Field) error {
	schemas := make([]tsapi.Field, len(fields))
	for i, f := range fields {
		schemas[i] = f.schema()
	}
	return cl.updateFields(ctx, name, schemas)
}

// DropFields removes fields, and their indexed data, from an existing collection.
func (cl *Client) DropFields(ctx context.Context, name string, fieldNames ...string) error {
	schemas := make([]tsapi.Field, len(fieldNames))
	for i, fieldName := range fieldNames {
		schemas[i] = tsapi.Field{Name: fieldName, Drop: new(true)}
	}
	return cl.updateFields(ctx, name, schemas)
}

func (cl *Client) updateFields(ctx context.Context, name string, fields []tsapi.Field) error {
	_, err := cl.ts.Collection(name).Update(ctx, &tsapi.CollectionUpdateSchema{Fields: fields})
	return err
}

// DeleteCollection removes the collection and every document in it.
func (cl *Client) DeleteCollection(ctx context.Context, name string) error {
	_, err := cl.ts.Collection(name).Delete(ctx)
	return err
}
