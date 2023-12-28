package validate

import (
	"strconv"
	"strings"
)

func Multi(fieldErrs ...*FieldError) error {
	if len(fieldErrs) == 0 {
		return nil
	}

	var err *Error
	err = err.Add(fieldErrs...)
	if err.Len() == 0 {
		return nil
	} else {
		return err
	}
}

func Required(field, value string) *FieldError {
	if strings.TrimSpace(value) == "" {
		return NewFieldError(field, "missing", value)
	} else {
		return nil
	}
}

func StrLen(field, value string, minLen int) *FieldError {
	if len(strings.TrimSpace(value)) < minLen {
		return NewFieldError(field, "too short, "+strconv.Itoa(minLen)+" characters required", value)
	} else {
		return nil
	}
}

func FieldsMatch(field1, value1, field2, value2 string) *FieldError {
	if value1 != value2 {
		return NewFieldError(field2, "does not match", value1)
	} else {
		return nil
	}
}
