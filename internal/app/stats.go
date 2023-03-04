// Package app can get stats
package app

import (
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/backend"
)

// Stats will retrieve entry stats
func Stats(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("entry required")
	}
	entry := args[0]
	v, err := cmd.Transaction().Get(entry, backend.StatsValue)
	if err != nil {
		return fmt.Errorf("unable to get stats: %w", err)
	}
	if v != nil {
		fmt.Fprintln(cmd.Writer(), v.Value)
	}
	return nil
}
