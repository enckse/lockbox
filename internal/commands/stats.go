// Package commands can get stats
package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/enckse/lockbox/internal/backend"
)

// Stats will retrieve entry stats
func Stats(w io.Writer, t *backend.Transaction, args []string) error {
	if len(args) != 1 {
		return errors.New("entry required")
	}
	entry := args[0]
	v, err := t.Get(entry, backend.StatsValue)
	if err != nil {
		return fmt.Errorf("unable to get stats: %w", err)
	}
	if v != nil {
		fmt.Fprintln(w, v.Value)
	}
	return nil
}
