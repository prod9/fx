package errutil

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}

func Validation(field, msg string) *ValidationError {
	return &ValidationError{field, msg, nil}
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func (e *ValidationError) WithMessage(msg string) *ValidationError {
	clone := *e
	clone.Message = msg
	return &clone
}

type ValidationErrors struct {
	errors []ValidationError
}

func (e ValidationErrors) AddError(err ValidationError) ValidationErrors {
	e.errors = append(e.errors, err)
	return e
}

func (e ValidationErrors) FieldMap() (res map[string]interface{}) {
	res = map[string]interface{}{}
	for _, err := range e.errors {
		res[err.Field] = err
	}
	return
}

func (e ValidationErrors) Size() int {
	return len(e.errors)
}

func (e ValidationErrors) Error() string {
	return "input validation error"
}
