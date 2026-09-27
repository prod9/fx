package typesense

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

// IsNotFound reports whether Typesense answered 404, e.g. for a missing collection.
func IsNotFound(err error) bool { return hasStatus(err, http.StatusNotFound) }

// IsConflict reports whether Typesense answered 409, e.g. when creating a collection
// whose name is taken.
func IsConflict(err error) bool { return hasStatus(err, http.StatusConflict) }

func hasStatus(err error, status int) bool {
	httpErr := &ts.HTTPError{}
	return errors.As(err, &httpErr) && httpErr.Status == status
}

// Raw exposes the underlying typesense-go client for operations this package does not
// wrap.
func (cl *Client) Raw() *ts.Client { return cl.ts }

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
