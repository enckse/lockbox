// Package app can remove an entry
package app

import (
	"errors"
	"fmt"

	"github.com/seanenck/lockbox/internal/backend"
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
	if len(existings) == 0 {
		return fmt.Errorf("no entities matching: %s", deleting)
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
		removals := func() []backend.TransactionEntity {
			var tx []backend.TransactionEntity
			for _, e := range existings {
				tx = append(tx, e.Transaction())
			}
			return tx
		}()
		if err := t.RemoveAll(removals); err != nil {
			return fmt.Errorf("unable to remove: %w", err)
		}
	}
	return nil
}
