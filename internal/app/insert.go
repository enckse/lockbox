// Package app can insert
package app

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/backend"
)

// Insert will execute an insert
func Insert(cmd UserInputOptions) error {
	t := cmd.Transaction()
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("invalid insert, no entry given")
	}
	entry := args[0]
	base := backend.Base(entry)
	if !slices.ContainsFunc(backend.AllowedFields, func(v string) bool {
		return base == strings.ToLower(v)
	}) {
		return fmt.Errorf("'%s' is not an allowed field name", base)
	}

	dir := backend.Directory(entry)
	existing, err := t.Get(dir, backend.SecretValue)
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
	password, err := cmd.Input(!isPipe && !strings.EqualFold(base, backend.NotesField))
	if err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	vals := make(backend.EntityValues)
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
