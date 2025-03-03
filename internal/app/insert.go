// Package app can insert
package app

import (
	"errors"
	"fmt"
	"strings"

	"codeberg.org/enckse/lockbox/internal/backend"
)

type (
	// InsertMode changes how inserts are handled
	InsertMode uint
)

const (
	// SingleLineInsert is a single line entry
	SingleLineInsert InsertMode = iota
	// MultiLineInsert is a multiline insert
	MultiLineInsert
	// TOTPInsert is a singleline but from TOTP subcommands
	TOTPInsert
)

// Insert will execute an insert
func Insert(cmd UserInputOptions, mode InsertMode) error {
	t := cmd.Transaction()
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("invalid insert, no entry given")
	}
	entry := args[0]
	existing, err := t.Get(entry, backend.BlankValue)
	if err != nil {
		return err
	}
	isPipe := cmd.IsPipe()
	if existing != nil {
		if !isPipe {
			if !cmd.Confirm("overwrite existing") {
				return nil
			}
		}
	}
	password, err := cmd.Input(!isPipe && mode != MultiLineInsert)
	if err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	p := strings.TrimSpace(string(password))
	if err := t.Insert(entry, p); err != nil {
		return err
	}
	if !isPipe {
		fmt.Fprintln(cmd.Writer())
	}
	return nil
}
