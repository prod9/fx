package validate

import (
	"strconv"
	"strings"
	"time"
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

		if valErr, ok := err.(*Error); ok {
			for _, fieldErrs := range valErr.Fields {
				outerr = outerr.Add(fieldErrs...)
			}
		} else if fieldErr, ok := err.(*FieldError); ok {
			outerr = outerr.Add(fieldErr)
		} else {
			panic("validate.Multi: errors must all be *Error or *FieldError")
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

func TimeRequired(field string, value time.Time) error {
	if value.IsZero() {
		return NewFieldError(field, "missing", value)
	} else {
		return nil
	}
}

func TimeBefore(field string, value time.Time, field2 string, value2 time.Time) error {
	if !value.Before(value2) {
		return NewFieldError(field, "must be before "+field2, value)
	} else {
		return nil
	}
}

func TimeAfter(field string, value time.Time, field2 string, value2 time.Time) error {
	if !value.After(value2) {
		return NewFieldError(field, "must be after "+field2, value)
	} else {
		return nil
	}
}
