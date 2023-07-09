// Package app can show/clip an entry
package app

import (
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/platform"
)

// ShowClip will handle showing/clipping an entry
func ShowClip(cmd CommandOptions, isShow bool) error {
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("only one argument supported")
	}
	entry := args[0]
	clipboard := platform.Clipboard{}
	if !isShow {
		var err error
		clipboard, err = platform.NewClipboard()
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
