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

func Marshal(entities []ExportEntity) ([]byte, error) {
	return json.MarshalIndent(entities, "", "    ")
}
