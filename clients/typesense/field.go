package typesense

import tsapi "github.com/typesense/typesense-go/v3/typesense/api"

// Field describes one field of a Collection. Every zero value keeps Typesense's own
// default, so only the settings that differ need to be written out.
type Field struct {
	Name     string
	Type     Type
	Optional bool
	NoIndex  bool   // Typesense indexes every field unless told otherwise
	Infix    bool   // enables infix (substring) search on the field
	Locale   string // "" keeps the server default
}

func (f Field) schema() tsapi.Field {
	field := tsapi.Field{Name: f.Name, Type: f.Type.String()}
	if f.Optional {
		field.Optional = new(true)
	}
	if f.NoIndex {
		field.Index = new(false)
	}
	if f.Infix {
		field.Infix = new(true)
	}
	if f.Locale != "" {
		field.Locale = new(f.Locale)
	}
	return field
}
