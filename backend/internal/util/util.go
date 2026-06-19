package util

import (
	"fmt"
	. "github.com/go-jet/jet/v2/postgres"
	"reflect"
)

func PrintStructFields(v interface{}, prefix string) {
	val := reflect.ValueOf(v)
	typ := val.Type()

	// Handle pointer to struct
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	// Ensure the input is a struct
	if typ.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := typ.Field(i)
			fieldName := fieldType.Name

			// Check if the field is an embedded struct
			if fieldType.Anonymous {
				// Recursive call for embedded struct
				PrintStructFields(field.Interface(), prefix+fieldName+".")
			} else {
				fmt.Printf("%s%s: %v\n", prefix, fieldName, field.Interface())
			}
		}
	} else {
		fmt.Println("Provided value is not a struct")
	}
}

func IntsToExpressions(ints []int) []Expression {
	expressions := make([]Expression, len(ints))
	for i, v := range ints {
		expressions[i] = Int(int64(v))
	}
	return expressions
}

func StringsToExpressions(strings []string) []Expression {
	expressions := make([]Expression, len(strings))
	for i, v := range strings {
		expressions[i] = String(v)
	}
	return expressions
}
