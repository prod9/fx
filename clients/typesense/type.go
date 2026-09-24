package typesense

import "fmt"

// Type is a Typesense field type.
// REF: https://typesense.org/docs/29.0/api/collections.html#field-types
type Type int

const (
	AutoType = Type(iota)
	StringType
	StringArrType
	StringAutoType
	Int32Type
	Int32ArrType
	Int64Type
	Int64ArrType
	FloatType
	FloatArrType
	BoolType
	BoolArrType
	GeopointType
	GeopointArrType
	ObjectType
	ObjectArrType
)

func (t Type) String() string {
	switch t {
	case AutoType:
		return "auto"
	case StringType:
		return "string"
	case StringArrType:
		return "string[]"
	case StringAutoType:
		return "string*"
	case Int32Type:
		return "int32"
	case Int32ArrType:
		return "int32[]"
	case Int64Type:
		return "int64"
	case Int64ArrType:
		return "int64[]"
	case FloatType:
		return "float"
	case FloatArrType:
		return "float[]"
	case BoolType:
		return "bool"
	case BoolArrType:
		return "bool[]"
	case GeopointType:
		return "geopoint"
	case GeopointArrType:
		return "geopoint[]"
	case ObjectType:
		return "object"
	case ObjectArrType:
		return "object[]"
	default:
		panic("typesense: unknown field type")
	}
}

// parseType maps a Typesense type name back to a Type. Types this package does not
// model, such as image, are errors rather than a silent guess.
func parseType(name string) (Type, error) {
	for t := AutoType; t <= ObjectArrType; t++ {
		if t.String() == name {
			return t, nil
		}
	}
	return AutoType, fmt.Errorf("typesense: unsupported field type %q", name)
}

func (t Type) isObject() bool { return t == ObjectType || t == ObjectArrType }
