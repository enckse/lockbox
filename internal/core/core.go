// Package core has helpers
package core

import (
	"fmt"
	"reflect"
	"sort"
)

func listFields[T SystemPlatform | JSONOutputMode](p any) []string {
	v := reflect.ValueOf(p)
	var vals []string
	for i := 0; i < v.NumField(); i++ {
		vals = append(vals, fmt.Sprintf("%v", v.Field(i).Interface().(T)))
	}
	sort.Strings(vals)
	return vals
}
