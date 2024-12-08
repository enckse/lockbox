// Package store is the internal memory store for loaded configuration settings
package store

import (
	"slices"
)

type (
	backing struct {
		integers map[string]int64
		strings  map[string]string
		booleans map[string]bool
		arrays   map[string][]string
	}

	// KeyValue are values exportable for interrogation beyond the store
	KeyValue struct {
		Key   string
		Value interface{}
	}
)

var configuration = newConfig()

func newConfig() backing {
	c := backing{}
	c.arrays = make(map[string][]string)
	c.integers = make(map[string]int64)
	c.booleans = make(map[string]bool)
	c.strings = make(map[string]string)
	return c
}

// Clear will clear store contents
func Clear() {
	configuration = newConfig()
}

// List will get the key/value list of settings
func List(filter ...string) []KeyValue {
	var results []KeyValue
	results = append(results, list(configuration.integers, GetInt64, filter)...)
	results = append(results, list(configuration.booleans, GetBool, filter)...)
	results = append(results, list(configuration.strings, GetString, filter)...)
	results = append(results, list(configuration.arrays, GetArray, filter)...)
	return results
}

func list[T any](m map[string]T, conv func(string) (T, bool), filter []string) []KeyValue {
	filtered := len(filter) > 0
	var result []KeyValue
	for k := range m {
		if filtered {
			if !slices.Contains(filter, k) {
				continue
			}
		}
		val, _ := conv(k)
		result = append(result, KeyValue{Key: k, Value: val})
	}
	return result
}

// GetInt64 will get an int64 value
func GetInt64(key string) (int64, bool) {
	return get(key, configuration.integers)
}

// GetBool will get a bool value
func GetBool(key string) (bool, bool) {
	return get(key, configuration.booleans)
}

// GetString will get a string value
func GetString(key string) (string, bool) {
	return get(key, configuration.strings)
}

// GetArray will get an array value
func GetArray(key string) ([]string, bool) {
	return get(key, configuration.arrays)
}

func get[T any](key string, m map[string]T) (T, bool) {
	val, ok := m[key]
	return val, ok
}

// SetInt64 will set an int64 value
func SetInt64(key string, val int64) {
	configuration.integers[key] = val
}

// SetBool will set a bool value
func SetBool(key string, val bool) {
	configuration.booleans[key] = val
}

// SetString will set a string value
func SetString(key, val string) {
	configuration.strings[key] = val
}

// SetArray will set an array value
func SetArray(key string, val []string) {
	configuration.arrays[key] = val
}
