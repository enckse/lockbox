// Package app can remove an entry
package app

import (
	"errors"
	"fmt"
	"io"

	"github.com/enckse/lockbox/internal/kdbx"
)

// Remove will remove an entry
func Remove(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("remove requires an entry")
	}
	return remove(cmd.Transaction(), cmd.Writer(), args[0], cmd)
}

func remove(t *kdbx.Transaction, w io.Writer, entry string, cmd CommandOptions) error {
	deleting := entry
	postfixRemove := "y"
	existings, err := t.MatchPath(deleting)
	if err != nil {
		return err
	}
	if len(existings) == 0 {
		return fmt.Errorf("no entities matching: %s", deleting)
	}
	if len(existings) > 1 {
		postfixRemove = "ies"
		fmt.Fprintln(w, "selected entities:")
		for _, e := range existings {
			fmt.Fprintf(w, " %s\n", e.Path)
		}
		fmt.Fprintln(w, "")
	}
	if cmd.Confirm(fmt.Sprintf("delete entr%s", postfixRemove)) {
		if err := t.RemoveAll(existings); err != nil {
			return fmt.Errorf("unable to remove: %w", err)
		}
	}
	return nil
}
