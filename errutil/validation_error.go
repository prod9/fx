package errutil

type ValidationError struct {
	Field   string
	Message string
}

func Validation(field, msg string) *ValidationError {
	return &ValidationError{field, msg}
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func (e *ValidationError) WithMessage(msg string) *ValidationError {
	clone := *e
	clone.Message = msg
	return &clone
}
