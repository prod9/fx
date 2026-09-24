package typesense

import "context"

// CreateCollection creates col on the server. There is deliberately no update: to change
// a schema, DestroyCollection the old one and CreateCollection the new one.
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

// DestroyCollection removes the collection and every document in it.
func (cl *Client) DestroyCollection(ctx context.Context, col Collection) error {
	_, err := cl.ts.Collection(col.Name).Delete(ctx)
	return err
}
