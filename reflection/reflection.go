package reflection

import (
	"fmt"
	"reflect"
	"strings"
)

// GetZeroValueOfType returns the zero value of a type
// if type is array it returns the zero value of the element type
func GetZeroValueOfType(value interface{}) reflect.Value {
	elementType := reflect.TypeOf(value).Elem()
	zeroVar := reflect.New(elementType).Elem()
	if elementType.Kind() != reflect.Pointer {
		zeroVar = reflect.New(reflect.PtrTo(elementType)).Elem()
	}
	return zeroVar
}

// CallMethod cass method of type by a variable 'value' and returns T
// panics if the method doesn't return T
func CallMethod[T any](value interface{}, name string) T {
	method := getMethodFromType(value, name)
	result := method.Call(nil)
	if len(result) > 0 {
		v, ok := result[0].Interface().(T)
		if !ok {
			panic(fmt.Sprintf("method '%s' of type '%s' doesnt return a '%T'", name, reflect.TypeOf(value).String(), *new(T)))
		}
		return v
	}
	return *new(T)
}

func getMethodFromType(value interface{}, name string) reflect.Value {
	typeName := reflect.TypeOf(value).String()
	valueOf := reflect.ValueOf(value)
	method := valueOf.MethodByName(name)

	if valueOf.Kind() == reflect.Slice {
		zeroVar := GetZeroValueOfType(valueOf.Interface())
		method = reflect.ValueOf(zeroVar.Interface()).MethodByName(name)
		typeName = zeroVar.Type().String()
	}

	if !method.IsValid() {
		panic(fmt.Sprintf("type '%s' doesn't implement method '%s'", typeName, name))
	}

	return method
}

// GetField will get a nested struct's field by dot notation string
func GetField[T any](obj interface{}, fieldPath string) (T, error) {
	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Split the field path by dot notation
	fields := strings.Split(fieldPath, ".")

	// Traverse through the fields
	for _, field := range fields {
		// Check if the value is valid and a struct
		if value.IsValid() && value.Kind() == reflect.Struct {
			fieldValue := value.FieldByName(field)

			// Check if the field exists
			if fieldValue.IsValid() {
				value = fieldValue
			} else {
				return *new(T), fmt.Errorf("field '%s' not found", field)
			}
		} else {
			return *new(T), fmt.Errorf("invalid value or non-struct type")
		}
	}

	// Return the final value
	res, _ := value.Interface().(T)
	return res, nil
}
