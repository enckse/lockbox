package app

import (
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/backend"
)

// ListFind will list/find entries
func ListFind(cmd CommandOptions, isFind bool) error {
	args := cmd.Args()
	opts := backend.QueryOptions{}
	opts.Mode = backend.ListMode
	if isFind {
		opts.Mode = backend.FindMode
		if len(args) < 1 {
			return errors.New("find requires search term")
		}
		opts.Criteria = args[0]
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
	for _, f := range e {
		fmt.Fprintf(w, "%s\n", f.Path)
	}
	return nil
}
