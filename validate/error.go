package validate

import "strings"

type Error struct {
	Fields map[string][]*FieldError `json:"fields"`
}

func (e *Error) Code() string   { return "validation" }
func (e *Error) ErrorData() any { return e.Fields }

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	msg := strings.Builder{}
	msg.WriteString("multiple validation errors:")
	for field, err := range e.Fields {
		for _, err := range err {
			msg.WriteString("\n" + field + ": " + err.Message)
		}
	}
	return msg.String()
}

func (e *Error) Len() int {
	if e == nil {
		return 0
	} else {
		return len(e.Fields)
	}
}

func (e *Error) Add(errs ...*FieldError) *Error {
	if e == nil {
		e = &Error{}
	}
	if e.Fields == nil {
		e.Fields = map[string][]*FieldError{}
	}

	for _, err := range errs {
		if err == nil {
			continue
		}

		e.Fields[err.Field] = append(e.Fields[err.Field], err)
	}

	if len(e.Fields) == 0 {
		return nil
	} else {
		return e
	}
}

func (e *Error) AddField(field, msg string, value any) *Error {
	return e.Add(NewFieldError(field, msg, value).(*FieldError))
}
