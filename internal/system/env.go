// Package system handles simple environment variable processing
package system

import (
	"os"
	"strings"
)

const (
	// Yes is the string expected for yes values
	Yes = "yes"
	// No is the string expected for no values
	No = "no"
)

type (
	// ReadValue is the output of reading a known bool/yes/no
	ReadValue uint
)

const (
	// UnknownValue indicates an unknown value was read
	UnknownValue ReadValue = iota
	// YesValue means yes was set
	YesValue
	// NoValue means no was set
	NoValue
	// EmptyValue means that the value was not set (empty string)
	EmptyValue
)

// EnvironOrDefault will get the environment value OR default if env is not set.
func EnvironOrDefault(envKey, defaultValue string) string {
	val := os.Getenv(envKey)
	if strings.TrimSpace(val) == "" {
		return defaultValue
	}
	return val
}

// EnvironValue read a simple yes/no from an environment value
func EnvironValue(envKey string) ReadValue {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(envKey)))
	switch value {
	case No:
		return NoValue
	case Yes:
		return YesValue
	case "":
		return EmptyValue
	}
	return UnknownValue
}
