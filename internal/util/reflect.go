// Package util has reflection helpers
package util

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type (
	// Position is the start/end of a word in a greater set
	Position struct {
		Start int
		End   int
	}
	// Word is the text and position in a greater position
	Word struct {
		Text     string
		Position Position
	}
)

// ListFields will get the values of strings on an "all string" struct
func ListFields(p any) []string {
	v := reflect.ValueOf(p)
	var vals []string
	for i := range v.NumField() {
		vals = append(vals, fmt.Sprintf("%v", v.Field(i).Interface()))
	}
	sort.Strings(vals)
	return vals
}

func readNested(v reflect.Type, root string) []string {
	var fields []string
	for i := range v.NumField() {
		field := v.Field(i)
		if field.Type.Kind() == reflect.Struct {
			fields = append(fields, readNested(field.Type, fmt.Sprintf("%s.", field.Name))...)
		} else {
			fields = append(fields, fmt.Sprintf("%s%s", root, field.Name))
		}
	}
	return fields
}

// TextPositionFields is the displayable set of templated fields
func TextPositionFields() string {
	return strings.Join(readNested(reflect.TypeOf(Word{}), ""), ", ")
}
