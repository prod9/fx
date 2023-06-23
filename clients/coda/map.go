package coda

import (
	"fmt"
	"reflect"
	"strings"
)

func Map(target any, row *Row) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	t := v.Type()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		return mapStruct(v, t, row)
		// TODO: Switch type here instead?
	}

	return fmt.Errorf("cannot map coda row to type %s", t.Name())
}

func mapStruct(v reflect.Value, t reflect.Type, row *Row) error {
	// TODO: Cache
	codaToField := map[string]int{}
	for idx := 0; idx < t.NumField(); idx++ {
		tag := strings.TrimSpace(t.Field(idx).Tag.Get("coda"))
		if tag != "" {
			codaToField[tag] = idx
		}
	}

	var err error
	for codaID, idx := range codaToField {
		field, fieldType, rowValue := v.Field(idx), t.Field(idx).Type, row.Values[codaID]
		switch fieldType.Kind() {
		case reflect.String:
			err = mapString(field, fieldType, rowValue)
		case reflect.Slice:
			err = mapSlice(field, fieldType, rowValue)
		case reflect.Bool:
			err = mapBool(field, fieldType, rowValue)
		default:
			err = fmt.Errorf("cannot map coda value to a `%s` field", field.Type())
		}

		if err != nil {
			return fmt.Errorf("error while mapping field `%s`: %w", t.Field(idx).Name, err)
		}
	}

	return nil
}

func mapString(v reflect.Value, t reflect.Type, value any) error {
	if str, ok := value.(string); ok {
		if strings.HasPrefix(str, "```") &&
			strings.HasSuffix(str, "```") {
			str = str[3 : len(str)-3]
		}

		v.Set(reflect.ValueOf(str))
		return nil

	} else if m, ok := value.(map[string]any); ok {
		switch m["@type"] {
		case "StructuredValue":
			v.Set(reflect.ValueOf(m["name"]))
			return nil
		case "WebPage":
			v.Set(reflect.ValueOf(m["url"]))
			return nil
		}
	}

	return fmt.Errorf("cannot map coda value `%#v` to string", value)
}

func mapSlice(v reflect.Value, t reflect.Type, value any) error {
	var slice []string
	if str, ok := value.(string); ok {
		slice = strings.Split(str, "\n- ")
		if strings.HasPrefix(slice[0], "- ") {
			slice[0] = slice[0][2:] // strip "- ", bullet list marker
		}

		v.Set(reflect.ValueOf(slice))
		return nil

	} else if codaSlice, ok := value.([]any); ok {
		for idx, codaItem := range codaSlice {
			str := ""
			if err := mapString(reflect.ValueOf(&str).Elem(), reflect.TypeOf(str), codaItem); err != nil {
				return fmt.Errorf("cannot map index %d: %w", idx, err)
			}
			slice = append(slice, str)
		}

		v.Set(reflect.ValueOf(slice))
		return nil

	}

	return fmt.Errorf("cannot map coda value `%#v` to string slice", value)
}

func mapBool(v reflect.Value, t reflect.Type, value any) error {
	if b, ok := value.(bool); ok {
		v.Set(reflect.ValueOf(b))
		return nil
	}
	// TODO: Strings and everything else

	return fmt.Errorf("cannot map coda value `%#v` to bool", value)
}
