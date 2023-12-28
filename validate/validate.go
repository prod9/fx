package validate

import (
	"strconv"
	"strings"
)

func Multi(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	var outerr *Error
	for _, err := range errs {
		if err == nil {
			continue
		}

		if fieldErr, ok := err.(*FieldError); !ok {
			panic("validate.Multi: errors must all be *FieldError")
		} else {
			outerr = outerr.Add(fieldErr)
		}
	}

	if outerr.Len() == 0 {
		return nil
	} else {
		return outerr
	}
}

func Required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return NewFieldError(field, "missing", value)
	} else {
		return nil
	}
}

func Positive(field string, value int64) error {
	if value <= 0 {
		return NewFieldError(field, "must be positive", value)
	} else {
		return nil
	}
}

func StrLen(field, value string, minLen int) error {
	if len(strings.TrimSpace(value)) < minLen {
		return NewFieldError(field, "too short, "+strconv.Itoa(minLen)+" characters required", value)
	} else {
		return nil
	}
}

func FieldsMatch(field1, value1, field2, value2 string) error {
	if value1 != value2 {
		return NewFieldError(field2, "does not match", value1)
	} else {
		return nil
	}
}
