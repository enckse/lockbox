// Package app can insert
package app

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/kdbx"
)

// Insert will execute an insert
func Insert(cmd UserInputOptions) error {
	t := cmd.Transaction()
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("invalid insert, no entry given")
	}
	entry := args[0]
	base := kdbx.Base(entry)
	if !slices.ContainsFunc(kdbx.AllowedFields, func(v string) bool {
		return base == strings.ToLower(v)
	}) {
		return fmt.Errorf("'%s' is not an allowed field name", base)
	}

	dir := kdbx.Directory(entry)
	existing, err := t.Get(dir, kdbx.SecretValue)
	if err != nil {
		return err
	}
	isPipe := cmd.IsPipe()
	if existing != nil {
		if !isPipe {
			if _, ok := existing.Value(base); ok {
				if !cmd.Confirm("overwrite existing") {
					return nil
				}
			}
		}
	}
	password, err := cmd.Input(!isPipe && !strings.EqualFold(base, kdbx.NotesField))
	if err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	vals := make(kdbx.EntityValues)
	if existing != nil {
		vals = existing.Values
	}
	vals[base] = strings.TrimSpace(string(password))
	if err := t.Insert(dir, vals); err != nil {
		return err
	}
	if !isPipe {
		fmt.Fprintln(cmd.Writer())
	}
	return nil
}
