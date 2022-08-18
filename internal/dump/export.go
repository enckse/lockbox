// Package dump handles export lockbox definitions to other formats.
package dump

import (
	"encoding/json"
)

type (
	// ExportEntity represents the output structure from a JSON dump.
	ExportEntity struct {
		Path  string `json:"path,omitempty"`
		Value string `json:"value"`
	}
)

// Marshal handles marshalling of entities to output formats.
func Marshal(entities []ExportEntity) ([]byte, error) {
	return json.MarshalIndent(entities, "", "    ")
}
