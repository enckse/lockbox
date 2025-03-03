// Package app can show/clip an entry
package app

import (
	"errors"
	"fmt"

	"git.sr.ht/~enckse/lockbox/internal/backend"
	"git.sr.ht/~enckse/lockbox/internal/platform/clip"
)

// ShowClip will handle showing/clipping an entry
func ShowClip(cmd CommandOptions, isShow bool) error {
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("only one argument supported")
	}
	entry := args[0]
	clipboard := clip.Board{}
	if !isShow {
		var err error
		clipboard, err = clip.New()
		if err != nil {
			return fmt.Errorf("unable to get clipboard: %w", err)
		}
	}
	existing, err := cmd.Transaction().Get(entry, backend.SecretValue)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("entry does not exist")
	}
	if isShow {
		fmt.Fprintln(cmd.Writer(), existing.Value)
		return nil
	}
	if err := clipboard.CopyTo(existing.Value); err != nil {
		return fmt.Errorf("clipboard operation failed: %w", err)
	}
	return nil
}
