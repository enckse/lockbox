// Package inputs handles user inputs/UI elements for JSON
package inputs

import (
	"fmt"
	"strings"

	"github.com/enckse/pgl/os/env"
)

type (
	// JSONOutputMode is the output mode definition
	JSONOutputMode uint
)

const (
	// JSONBlankMode will results in an empty field
	JSONBlankMode JSONOutputMode = iota
	// JSONHashMode will use a common hasher to hash the raw value
	JSONHashMode
	// JSONRawMode will display the raw value
	JSONRawMode
)

// ParseJSONOutput handles detecting the JSON output mode
func ParseJSONOutput() (JSONOutputMode, error) {
	val := strings.ToLower(strings.TrimSpace(env.GetOrDefault(JSONDataOutputEnv, JSONDataOutputHash)))
	switch val {
	case JSONDataOutputHash:
		return JSONHashMode, nil
	case JSONDataOutputBlank:
		return JSONBlankMode, nil
	case JSONDataOutputRaw:
		return JSONRawMode, nil
	}
	return JSONBlankMode, fmt.Errorf("invalid JSON output mode: %s", val)
}
