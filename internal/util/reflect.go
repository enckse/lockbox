// Package util has reflection helpers
package util

import (
	"fmt"
	"reflect"
	"sort"
)

// ListFields will get the values of strings on an "all string" struct
func ListFields(p any) []string {
	v := reflect.ValueOf(p)
	var vals []string
	for i := 0; i < v.NumField(); i++ {
		vals = append(vals, fmt.Sprintf("%v", v.Field(i).Interface()))
	}
	sort.Strings(vals)
	return vals
}
