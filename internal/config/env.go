// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/enckse/lockbox/internal/config/store"
)

type (
	environmentBase struct {
		key         string
		description string
		requirement string
	}
	environmentDefault[T any] struct {
		value T
		environmentBase
	}
	environmentStrings struct {
		environmentDefault[string]
		flags   []stringsFlags
		allowed []string
	}
	// EnvironmentInt are environment settings that are integers
	EnvironmentInt struct {
		environmentDefault[int]
		short   string
		canZero bool
	}
	// EnvironmentBool are environment settings that are booleans
	EnvironmentBool struct {
		environmentDefault[bool]
	}
	// EnvironmentFormatter allows for sending a string into a get request
	EnvironmentFormatter struct {
		fxn func(string, string) string
		environmentBase
		allowed string
	}
	// EnvironmentString represents a string variable
	EnvironmentString struct {
		environmentStrings
	}
	// EnvironmentArray is an array of strings variable
	EnvironmentArray struct {
		environmentStrings
	}
	metaData struct {
		value     string
		tomlType  tomlType
		tomlValue string
		allowed   []string
		canExpand bool
	}
)

func (e environmentBase) Key() string {
	return fmt.Sprintf(environmentPrefix+"%s", e.key)
}

// Get will get the boolean value for the setting
func (e EnvironmentBool) Get() bool {
	val, ok := store.GetBool(e.Key())
	if !ok {
		val = e.value
	}
	return val
}

// Get will get the integer value for the setting
func (e EnvironmentInt) Get() (int64, error) {
	i, ok := store.GetInt64(e.Key())
	if !ok {
		i = int64(e.value)
	}
	invalid := false
	check := ""
	if e.canZero {
		check = "="
	}
	switch i {
	case 0:
		invalid = !e.canZero
	default:
		invalid = i < 0
	}
	if invalid {
		return -1, fmt.Errorf("%s must be >%s 0", e.short, check)
	}
	return i, nil
}

// Get will read the string from the environment
func (e EnvironmentString) Get() string {
	return stringsGet(e.environmentStrings, store.GetString, func(val string) string {
		return val
	})
}

// Get indicates the item should be queried as an array
func (e EnvironmentArray) Get() []string {
	return stringsGet(e.environmentStrings, store.GetArray, func(val string) []string {
		return strings.Split(val, arrayDelimiter)
	})
}

func stringsGet[T string | []string](e environmentStrings, getter func(string) (T, bool), conv func(string) T) T {
	val, ok := getter(e.Key())
	if !ok {
		flags := e.flattenFlags()
		if slices.Contains(flags, canDefaultFlag) {
			val = conv(e.value)
		}
	}
	return val
}

// Get will retrieve the value with the formatted input included
func (e EnvironmentFormatter) Get(value string) string {
	return e.fxn(e.Key(), value)
}

func (e EnvironmentArray) display() metaData {
	return e.environmentStrings.display(true)
}

func (e EnvironmentString) display() metaData {
	return e.environmentStrings.display(false)
}

func (e environmentStrings) display(isArray bool) metaData {
	var t tomlType
	t = tomlString
	v := "\"\""
	show := e.allowed
	value := e.value
	flags := e.flattenFlags()
	canExpand := slices.Contains(flags, canExpandFlag)
	if isArray {
		t = tomlArray
		v = "[]"
		if slices.Contains(flags, isCommandFlag) {
			if len(show) == 0 {
				show = []string{"[cmd args...]"}
			}
			if value == "" {
				value = "(detected)"
			}
		}
	}
	return metaData{
		value:     value,
		allowed:   show,
		tomlType:  t,
		tomlValue: v,
		canExpand: canExpand,
	}
}

func (e environmentBase) self() environmentBase {
	return e
}

func (e EnvironmentBool) display() metaData {
	val := NoValue
	if e.value {
		val = YesValue
	}
	return metaData{
		value:     val,
		allowed:   []string{YesValue, NoValue},
		tomlType:  tomlBool,
		tomlValue: YesValue,
		canExpand: false,
	}
}

func (e EnvironmentInt) display() metaData {
	return metaData{
		value:     fmt.Sprintf("%d", e.value),
		allowed:   []string{"<integer>"},
		tomlType:  tomlInt,
		tomlValue: "0",
		canExpand: false,
	}
}

func (e EnvironmentFormatter) display() metaData {
	return metaData{
		value:     strings.ReplaceAll(strings.ReplaceAll(EnvTOTPFormat.Get("%s"), "%25s", "%s"), "&", " \\\n           &"),
		allowed:   []string{e.allowed},
		tomlType:  tomlString,
		tomlValue: "\"\"",
		canExpand: false,
	}
}

func (e environmentStrings) flattenFlags() []stringsFlags {
	flags := e.flags
	if slices.Contains(e.flags, isCommandFlag) {
		flags = append(flags, canExpandFlag)
	}
	return flags
}
