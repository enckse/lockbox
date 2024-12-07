// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
func (e EnvironmentBool) Get() (bool, error) {
	return parseStringYesNo(e, getExpand(e.Key()))
}

func parseStringYesNo(e EnvironmentBool, in string) (bool, error) {
	read := strings.ToLower(strings.TrimSpace(in))
	switch read {
	case no:
		return false, nil
	case yes:
		return true, nil
	case "":
		return e.defaultValue, nil
	}

	return false, fmt.Errorf("invalid yes/no env value for %s", e.Key())
}

// Get will get the integer value for the setting
func (e EnvironmentInt) Get() (int, error) {
	val := e.defaultValue
	use := getExpand(e.Key())
	if use != "" {
		i, err := strconv.Atoi(use)
		if err != nil {
			return -1, err
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
		val = i
	}
	return val, nil
}

// Get will read the string from the environment
func (e EnvironmentString) Get() string {
	if !e.canDefault {
		return getExpand(e.Key())
	}
	return environOrDefault(e.Key(), e.defaultValue)
}

// Get will read (and shlex) the value if set
func (e EnvironmentCommand) Get() ([]string, error) {
	value := environOrDefault(e.Key(), "")
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	return shlex(value)
}

// KeyValue will get the string representation of the key+value
func (e environmentBase) KeyValue(value string) string {
	return fmt.Sprintf("%s=%s", e.Key(), value)
}

// Setenv will do an environment set for the value to key
func (e environmentBase) Set(value string) error {
	unset, err := IsUnset(e.Key(), value)
	if err != nil {
		return err
	}
	if unset {
		return nil
	}
	return os.Setenv(e.Key(), value)
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
	val := no
	if e.defaultValue {
		val = yes
	}
	return val, []string{yes, no}
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
