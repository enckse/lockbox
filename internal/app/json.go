// Package app can get stats
package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/backend"
)

// JSON will get entries (1 or ALL) in JSON format
func JSON(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) > 1 {
		return errors.New("invalid arguments")
	}
	if len(args) == 0 {
		return serialize(cmd.Writer(), cmd.Transaction())
	}
	entry := args[0]
	v, err := cmd.Transaction().Get(entry, backend.JSONValue)
	if err != nil {
		return fmt.Errorf("unable to get json: %w", err)
	}
	if v != nil {
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(v.Value), "", "  "); err != nil {
			return err
		}
		fmt.Fprintln(cmd.Writer(), buf.String())
	}
	return nil
}
