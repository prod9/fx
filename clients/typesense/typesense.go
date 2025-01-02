package typesense

import (
	"context"
	"errors"
	"fmt"

	"fx.prodigy9.co/config"
	"github.com/typesense/typesense-go/v2/typesense"
	ts "github.com/typesense/typesense-go/v2/typesense"
	tsapi "github.com/typesense/typesense-go/v2/typesense/api"
)

var (
	ServerConfig = config.StrDef("TYPESENSE_SERVER", "http://localhost:8108")
	APIKeyConfig = config.Str("TYPESENSE_API_KEY")
)

func IsNotFound(err error) bool {
	httpErr := &typesense.HTTPError{}
	if errors.As(err, &httpErr) {
		return httpErr.Status == 404
	} else {
		return false
	}
}

type Client struct {
	ts *ts.Client
}

func New(cfg *config.Source) *Client {
	return &Client{ts.NewClient(
		ts.WithServer(config.Get(cfg, ServerConfig)),
		ts.WithAPIKey(config.Get(cfg, APIKeyConfig)),
	)}
}

func (cl *Client) CreateCollection(ctx context.Context, col Collection) error {
	_, err := cl.ts.Collections().Create(ctx, &col.impl().schema)
	return err
}
func (cl *Client) DestroyCollection(ctx context.Context, col Collection) error {
	_, err := cl.ts.Collection(col.Name()).Delete(ctx)
	return err
}
func (cl *Client) Index(ctx context.Context, col Collection, obj any) error {
	_, err := cl.ts.Collection(col.Name()).Documents().Upsert(ctx, obj)
	return err
}

func (cl *Client) Search(ctx context.Context, col Collection, field, q string, out any) error {
	result, err := cl.ts.Collection(col.Name()).Documents().Search(ctx,
		&tsapi.SearchCollectionParams{
			Q:       &q,
			QueryBy: &field,
		})
	if err != nil {
		return err
	}

	for _, hit := range *result.Hits {
		fmt.Printf("%#v\n", *hit.Document)
	}
	return nil
}
