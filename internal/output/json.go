// Package output defines JSON settings/modes
package output

import (
	"fmt"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/util"
)

// JSONModes are the JSON data output types for exporting/output of values
var JSONModes = JSONTypes{
	Hash:  "hash",
	Blank: "empty",
	Raw:   "plaintext",
}

type (
	// JSONMode is the output mode definition
	JSONMode string

	// JSONTypes indicate how JSON data can be exported for values
	JSONTypes struct {
		Hash  JSONMode
		Blank JSONMode
		Raw   JSONMode
	}
)

// List will list the output modes on the struct
func (p JSONTypes) List() []string {
	return util.ListFields(p)
}

// ParseJSONMode handles detecting the JSON output mode
func ParseJSONMode(value string) (JSONMode, error) {
	val := JSONMode(strings.ToLower(strings.TrimSpace(value)))
	switch val {
	case JSONModes.Hash, JSONModes.Blank, JSONModes.Raw:
		return val, nil
	}
	return JSONModes.Blank, fmt.Errorf("invalid JSON output mode: %s", val)
}
