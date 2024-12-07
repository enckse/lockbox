// Package core defines JSON outputs
package core

import (
	"fmt"
	"strings"
)

// JSONOutputs are the JSON data output types for exporting/output of values
var JSONOutputs = JSONOutputTypes{
	Hash:  "hash",
	Blank: "empty",
	Raw:   "plaintext",
}

type (
	// JSONOutputMode is the output mode definition
	JSONOutputMode string

	// JSONOutputTypes indicate how JSON data can be exported for values
	JSONOutputTypes struct {
		Hash  JSONOutputMode
		Blank JSONOutputMode
		Raw   JSONOutputMode
	}
)

// List will list the output modes on the struct
func (p JSONOutputTypes) List() []string {
	return listFields[JSONOutputMode](p)
}

// ParseJSONOutput handles detecting the JSON output mode
func ParseJSONOutput(value string) (JSONOutputMode, error) {
	val := JSONOutputMode(strings.ToLower(strings.TrimSpace(value)))
	switch val {
	case JSONOutputs.Hash, JSONOutputs.Blank, JSONOutputs.Raw:
		return val, nil
	}
	return JSONOutputs.Blank, fmt.Errorf("invalid JSON output mode: %s", val)
}
