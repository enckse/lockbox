// Package inputs handles user inputs/UI elements for JSON
package inputs

import (
	"fmt"
	"strings"
)

const (
	// JSONDataOutputHash means output data is hashed
	JSONDataOutputHash JSONOutputMode = "hash"
	// JSONDataOutputBlank means an empty entry is set
	JSONDataOutputBlank JSONOutputMode = "empty"
	// JSONDataOutputRaw means the RAW (unencrypted) value is displayed
	JSONDataOutputRaw JSONOutputMode = "plaintext"
)

type (
	// JSONOutputMode is the output mode definition
	JSONOutputMode string
)

// ParseJSONOutput handles detecting the JSON output mode
func ParseJSONOutput() (JSONOutputMode, error) {
	val := strings.ToLower(strings.TrimSpace(EnvironOrDefault(JSONDataOutputEnv, string(JSONDataOutputHash))))
	switch JSONOutputMode(val) {
	case JSONDataOutputHash:
		return JSONDataOutputHash, nil
	case JSONDataOutputBlank:
		return JSONDataOutputBlank, nil
	case JSONDataOutputRaw:
		return JSONDataOutputRaw, nil
	}
	return JSONDataOutputBlank, fmt.Errorf("invalid JSON output mode: %s", val)
}
