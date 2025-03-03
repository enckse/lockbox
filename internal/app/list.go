// Package app can list entries
package app

import (
	"errors"
	"fmt"

	"codeberg.org/enckse/lockbox/internal/backend"
)

// List will list/find entries
func List(cmd CommandOptions) error {
	args := cmd.Args()
	opts := backend.QueryOptions{}
	opts.Mode = backend.ListMode
	if len(args) != 0 {
		return errors.New("list does not support any arguments")
	}
	e, err := cmd.Transaction().QueryCallback(opts)
	if err != nil {
		return err
	}
	w := cmd.Writer()
	for f, err := range e {
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\n", f.Path)
	}
	return nil
}
