// Package app can remove an entry
package app

import (
	"errors"
	"fmt"
)

// Remove will remove an entry
func Remove(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("remove requires an entry")
	}
	t := cmd.Transaction()
	deleting := args[0]
	postfixRemove := "y"
	existings, err := t.MatchPath(deleting)
	if err != nil {
		return err
	}
	w := cmd.Writer()
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
