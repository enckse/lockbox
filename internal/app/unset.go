// Package app can unset a field
package app

import (
	"errors"
	"fmt"

	"git.sr.ht/~enckse/lockbox/internal/kdbx"
)

// Unset enables clearing an entry
func Unset(cmd CommandOptions) error {
	t := cmd.Transaction()
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("invalid unset, no entry given")
	}
	entry := args[0]
	base := kdbx.Base(entry)
	dir := kdbx.Directory(entry)
	existing, err := t.Get(dir, kdbx.SecretValue)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("%s does not exist", entry)
	}
	w := cmd.Writer()
	unsetRemove := func(v kdbx.EntityValues) (bool, error) {
		if len(v) == 0 {
			fmt.Fprintf(w, "removing empty group: %s\n", dir)
			return true, remove(t, w, dir, cmd)
		}
		return false, nil
	}
	ok, err := unsetRemove(existing.Values)
	if ok {
		return err
	}
	if _, ok := existing.Value(base); ok {
		vals := existing.Values
		delete(vals, base)
		ok, err = unsetRemove(vals)
		if ok {
			return err
		}
		if !cmd.Confirm(fmt.Sprintf("unset: %s", entry)) {
			return nil
		}
		fmt.Fprintf(w, "clearing value from: %s\n", entry)
		if err := t.Insert(dir, vals); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unable to unset: %s", entry)
}
