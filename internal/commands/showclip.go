// Package commands can show/clip an entry
package commands

import (
	"errors"
	"fmt"
	"io"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/platform"
)

// ShowClip will handle showing/clipping an entry
func ShowClip(w io.Writer, t *backend.Transaction, isShow bool, args []string) error {
	if len(args) != 1 {
		return errors.New("entry required")
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
	existing, err := t.Get(entry, backend.SecretValue)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	if isShow {
		fmt.Fprintln(w, existing.Value)
		return nil
	}
	if err := clipboard.CopyTo(existing.Value); err != nil {
		return fmt.Errorf("clipboard operation failed: %w", err)
	}
	return nil
}
