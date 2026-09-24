package typesense

import (
	"slices"

	tsapi "github.com/typesense/typesense-go/v3/typesense/api"
)

// Collection is a Typesense collection schema.
type Collection struct {
	Name        string
	Fields      []Field
	DefaultSort string // "" means no default sorting field
}

func (c Collection) schema() *tsapi.CollectionSchema {
	fields := make([]tsapi.Field, len(c.Fields))
	for i, f := range c.Fields {
		fields[i] = f.schema()
	}

	schema := &tsapi.CollectionSchema{Name: c.Name, Fields: fields}
	if c.DefaultSort != "" {
		schema.DefaultSortingField = new(c.DefaultSort)
	}
	// Typesense rejects object fields unless nested fields are enabled, so enable
	// them whenever the schema declares one rather than making callers remember.
	if slices.ContainsFunc(c.Fields, func(f Field) bool { return f.Type.isObject() }) {
		schema.EnableNestedFields = new(true)
	}
	return schema
}
