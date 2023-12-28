package validate

type FieldError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func NewFieldError(field, msg string, value any) *FieldError {
	return &FieldError{field, msg, value}
}

func (e *FieldError) Code() string { return "validation" }
func (e *FieldError) ErrorData() any {
	// ErrorData is called if FieldError is rendered at top-level (no *Error parent) so we
	// match the structure with *Error so it's can be handled at the frontend in a
	// consistent way
	return map[string][]*FieldError{e.Field: {e}}
}

func (e *FieldError) Error() string {
	return e.Field + ": " + e.Message
}

func (e *FieldError) WithMessage(msg string) *FieldError {
	clone := *e
	clone.Message = msg
	return &clone
}
