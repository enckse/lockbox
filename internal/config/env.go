// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"strings"

	"github.com/seanenck/lockbox/internal/config/store"
)

type (
	environmentBase struct {
		subKey      string
		cat         keyCategory
		desc        string
		requirement string
	}
	environmentDefault[T any] struct {
		environmentBase
		defaultValue T
	}
	// EnvironmentInt are environment settings that are integers
	EnvironmentInt struct {
		environmentDefault[int]
		allowZero bool
		shortDesc string
	}
	// EnvironmentBool are environment settings that are booleans
	EnvironmentBool struct {
		environmentDefault[bool]
	}
	// EnvironmentString are string-based settings
	EnvironmentString struct {
		environmentDefault[string]
		canDefault bool
		allowed    []string
		isArray    bool
		expand     bool
	}
	// EnvironmentCommand are settings that are parsed as shell commands
	EnvironmentCommand struct {
		environmentBase
	}
	// EnvironmentFormatter allows for sending a string into a get request
	EnvironmentFormatter struct {
		environmentBase
		allowed string
		fxn     func(string, string) string
	}
)

func (e environmentBase) Key() string {
	return fmt.Sprintf(environmentPrefix+"%s%s", string(e.cat), e.subKey)
}

// Get will get the boolean value for the setting
func (e EnvironmentBool) Get() bool {
	val, ok := store.GetBool(e.Key())
	if !ok {
		val = e.defaultValue
	}
	return val
}

// Get will get the integer value for the setting
func (e EnvironmentInt) Get() (int64, error) {
	i, ok := store.GetInt64(e.Key())
	if !ok {
		i = int64(e.defaultValue)
	}
	invalid := false
	check := ""
	if e.allowZero {
		check = "="
	}
	switch i {
	case 0:
		invalid = !e.allowZero
	default:
		invalid = i < 0
	}
	if invalid {
		return -1, fmt.Errorf("%s must be >%s 0", e.shortDesc, check)
	}
	return i, nil
}

// Get will read the string from the environment
func (e EnvironmentString) Get() string {
	val, ok := store.GetString(e.Key())
	if !ok {
		if !e.canDefault {
			return ""
		}
		val = e.defaultValue
	}
	return val
}

// Get will read (and shlex) the value if set
func (e EnvironmentCommand) Get() []string {
	val, ok := store.GetArray(e.Key())
	if !ok {
		return []string{}
	}
	return val
}

// Get will retrieve the value with the formatted input included
func (e EnvironmentFormatter) Get(value string) string {
	return e.fxn(e.Key(), value)
}

func (e EnvironmentString) values() (string, []string) {
	return e.defaultValue, e.allowed
}

func (e environmentBase) self() environmentBase {
	return e
}

func (e EnvironmentBool) values() (string, []string) {
	val := NoValue
	if e.defaultValue {
		val = YesValue
	}
	return val, []string{YesValue, NoValue}
}

func (e EnvironmentInt) values() (string, []string) {
	return fmt.Sprintf("%d", e.defaultValue), []string{"<integer>"}
}

func (e EnvironmentFormatter) values() (string, []string) {
	return strings.ReplaceAll(strings.ReplaceAll(EnvTOTPFormat.Get("%s"), "%25s", "%s"), "&", " \\\n           &"), []string{e.allowed}
}

func (e EnvironmentCommand) values() (string, []string) {
	return detectedValue, []string{commandArgsExample}
}

func (e EnvironmentInt) toml() (tomlType, string, bool) {
	return tomlInt, "0", false
}

func (e EnvironmentBool) toml() (tomlType, string, bool) {
	return tomlBool, YesValue, false
}

func (e EnvironmentString) toml() (tomlType, string, bool) {
	if e.isArray {
		return tomlArray, "[]", e.expand
	}
	return tomlString, "\"\"", e.expand
}

func (e EnvironmentCommand) toml() (tomlType, string, bool) {
	return tomlArray, "[]", true
}

func (e EnvironmentFormatter) toml() (tomlType, string, bool) {
	return tomlString, "\"\"", false
}
