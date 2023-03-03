// Package commands can remove an entry
package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/enckse/lockbox/internal/backend"
)

// Remove will remove an entry
func Remove(w io.Writer, t *backend.Transaction, args []string, confirm Confirm) error {
	if len(args) != 1 {
		return errors.New("remove requires an entry")
	}
	deleting := args[0]
	postfixRemove := "y"
	existings, err := t.MatchPath(deleting)
	if err != nil {
		return err
	}

	if len(existings) > 1 {
		postfixRemove = "ies"
		fmt.Fprintln(w, "selected entities:")
		for _, e := range existings {
			fmt.Fprintf(w, " %s\n", e.Path)
		}
		fmt.Fprintln(w, "")
	}
	if confirm(fmt.Sprintf("delete entr%s", postfixRemove)) {
		if err := t.RemoveAll(existings); err != nil {
			return fmt.Errorf("unable to remove: %w", err)
		}
	}
	return nil
}
