// Package app can list entries
package app

import (
	"errors"
	"fmt"

	"git.sr.ht/~enckse/lockbox/internal/backend"
)

// List will list/find entries
func List(cmd CommandOptions, isFind bool) error {
	args := cmd.Args()
	opts := backend.QueryOptions{}
	opts.Mode = backend.ListMode
	if isFind {
		if len(args) != 1 {
			return errors.New("find requires one argument")
		}
		opts.PathFilter = args[0]
	} else {
		if len(args) != 0 {
			return errors.New("list does not support any arguments")
		}
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
