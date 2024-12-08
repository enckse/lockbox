package store

import (
	"slices"
)

type backing struct {
	integers map[string]int64
	strings  map[string]string
	booleans map[string]bool
	arrays   map[string][]string
	all      map[string]struct{}
}

type KeyValue struct {
	Key   string
	Value interface{}
}

var configuration = newConfig()

func newConfig() backing {
	c := backing{}
	c.arrays = make(map[string][]string)
	c.integers = make(map[string]int64)
	c.booleans = make(map[string]bool)
	c.strings = make(map[string]string)
	return c
}

func Clear() {
	configuration = newConfig()
}

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

func GetInt64(key string) (int64, bool) {
	return get(key, configuration.integers)
}

func GetBool(key string) (bool, bool) {
	return get(key, configuration.booleans)
}

func GetString(key string) (string, bool) {
	return get(key, configuration.strings)
}

func GetArray(key string) ([]string, bool) {
	return get(key, configuration.arrays)
}

func get[T any](key string, m map[string]T) (T, bool) {
	val, ok := m[key]
	return val, ok
}

func SetInt64(key string, val int64) {
	configuration.integers[key] = val
}

func SetBool(key string, val bool) {
	configuration.booleans[key] = val
}

func SetString(key string, val string) {
	configuration.strings[key] = val
}

func SetArray(key string, val []string) {
	configuration.arrays[key] = val
}
