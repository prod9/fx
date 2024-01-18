package resources

import (
	"context"
	"database/sql"
	"fmt"
	"fx.prodigy9.co/contrib/structs"
	"github.com/ggicci/httpin"
	"reflect"
)

type ProviderOptions struct {
	ID      string
	OwnerID string
	Order   string
}

type ResourceProvider func(ctx context.Context, opts ProviderOptions) interface{}

// MapResourcesFromRoute will use the provider map to get a resource from the database and set it in the input form
func MapResourcesFromRoute(ctx context.Context, providerMap map[string]ResourceProvider) error {
	input := ctx.Value(httpin.Input)
	refVal := reflect.ValueOf(input)
	parsed := structs.Parse(input)
	order := parsed.FindFieldByTag("fx", "order", nil)

	for _, field := range parsed.GetResourceFields() {
		provider, _ := providerMap[field.DbTable]
		if provider == nil {
			return fmt.Errorf("missing resource provider for table '%s'", field.DbTable)
		}

		opts := ProviderOptions{
			ID:      field.ID,
			OwnerID: field.OwnerID,
		}
		if order != nil {
			opts.Order, _ = order.Value.(string)
		}
		resource := provider(ctx, opts)
		if resource == nil {
			if field.IsRequired {
				return sql.ErrNoRows
			}
			continue
		}
		refVal.Elem().FieldByName(field.Name).Set(reflect.ValueOf(resource))
	}
	return nil
}

// Provider wraps the resource provider with common functionality
func Provider[T any](provider func(ctx context.Context, opts ProviderOptions) ([]T, error)) ResourceProvider {
	return func(ctx context.Context, opts ProviderOptions) interface{} {
		result, err := provider(ctx, opts)
		if err != nil {
			return nil
		}
		getSingleResource := opts.ID != "" && opts.OwnerID == "" || opts.ID != "" && opts.OwnerID != ""
		getList := opts.ID == "" && opts.OwnerID != ""

		if getSingleResource {
			if len(result) == 0 {
				return nil
			}
			return result[0]
		}
		if getList {
			return result
		}
		return nil
	}
}
