package typesense

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fx.prodigy9.co/config"
	ts "github.com/typesense/typesense-go/v3/typesense"
	tsapi "github.com/typesense/typesense-go/v3/typesense/api"
)

var (
	ServerConfig  = config.StrDef("TYPESENSE_SERVER", "http://localhost:8108")
	APIKeyConfig  = config.Str("TYPESENSE_API_KEY")
	TimeoutConfig = config.DurationDef("TYPESENSE_TIMEOUT", 5*time.Second)
)

type Client struct {
	ts *ts.Client
}

// New never fails: the underlying client connects lazily, so a misconfigured server
// surfaces as an error on the first call instead of blocking application startup.
func New(cfg *config.Source) *Client {
	return &Client{ts.NewClient(
		ts.WithServer(config.Get(cfg, ServerConfig)),
		ts.WithAPIKey(config.Get(cfg, APIKeyConfig)),
		ts.WithConnectionTimeout(config.Get(cfg, TimeoutConfig)),
	)}
}

func IsNotFound(err error) bool {
	httpErr := &ts.HTTPError{}
	if errors.As(err, &httpErr) {
		return httpErr.Status == 404
	} else {
		return false
	}
}

// Raw exposes the underlying typesense-go client for operations this package does not
// wrap.
func (cl *Client) Raw() *ts.Client { return cl.ts }

func (cl *Client) CreateCollection(ctx context.Context, col Collection) error {
	_, err := cl.ts.Collections().Create(ctx, col.schema())
	return err
}
func (cl *Client) DestroyCollection(ctx context.Context, col Collection) error {
	_, err := cl.ts.Collection(col.Name).Delete(ctx)
	return err
}
func (cl *Client) Index(ctx context.Context, col Collection, obj any) error {
	_, err := cl.ts.Collection(col.Name).Documents().Upsert(ctx, obj, &tsapi.DocumentIndexParameters{})
	return err
}

func (cl *Client) Search(ctx context.Context, col Collection, field, q string, out any) error {
	result, err := cl.ts.Collection(col.Name).Documents().Search(ctx,
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
