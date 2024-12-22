package typesense

// REF: https://typesense.org/docs/27.0/api/collections.html#field-types
type Type int

const (
	AutoType = Type(iota)
	StringType
	StringArrType
	Int32Type
	Int32ArrType
	Int64Type
	Int64ArrType
	FloatType
	FloatArrType
)

func (t Type) String() string {
	switch t {
	case AutoType:
		return "auto"
	case StringType:
		return "string"
	case StringArrType:
		return "string[]"
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

	default:
		// panic?
		return ""
	}
}
