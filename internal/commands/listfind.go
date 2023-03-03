package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
)

// ListFind will list/find entries
func ListFind(t *backend.Transaction, w io.Writer, command string, args []string) error {
	opts := backend.QueryOptions{}
	opts.Mode = backend.ListMode
	if command == cli.FindCommand {
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
	e, err := t.QueryCallback(opts)
	if err != nil {
		return err
	}
	for _, f := range e {
		fmt.Fprintf(w, "%s\n", f.Path)
	}
	return nil
}
