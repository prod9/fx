package typesense

import tsapi "github.com/typesense/typesense-go/v2/typesense/api"

type (
	Collection interface {
		Name() string
		impl() *collectionImpl
	}
	CollectionBuilder interface {
		SortOn(name string) CollectionBuilder
		Field(name string, typ Type, optional, facet bool) CollectionBuilder
		Build() Collection
	}

	collectionImpl struct{ schema tsapi.CollectionSchema }
	builderImpl    struct{ schema tsapi.CollectionSchema }
)

func (c *collectionImpl) Name() string          { return c.schema.Name }
func (c *collectionImpl) impl() *collectionImpl { return c }

func BuildCollection(name string) CollectionBuilder {
	impl := &builderImpl{}
	impl.schema.Name = name
	return impl
}

func (b *builderImpl) SortOn(name string) CollectionBuilder {
	b.schema.DefaultSortingField = &name
	return b
}
func (b *builderImpl) Field(name string, typ Type, optional, facet bool) CollectionBuilder {
	b.schema.Fields = append(b.schema.Fields, tsapi.Field{
		Name:     name,
		Type:     typ.String(),
		Optional: &optional,
		Facet:    &facet,
	})
	return b
}

func (b *builderImpl) Build() Collection {
	return &collectionImpl{b.schema}
}
