// Package app can show/clip an entry
package app

import (
	"errors"
	"fmt"

	"git.sr.ht/~enckse/lockbox/internal/kdbx"
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
		clipboard, err = clip.New(clip.DefaultLoader{Full: false})
		if err != nil {
			return fmt.Errorf("unable to get clipboard: %w", err)
		}
	}
	val, err := getEntity(entry, cmd)
	if err != nil {
		return err
	}
	if isShow {
		fmt.Fprintln(cmd.Writer(), val)
		return nil
	}
	if err := clipboard.CopyTo(val); err != nil {
		return fmt.Errorf("clipboard operation failed: %w", err)
	}
	return nil
}

func getEntity(entry string, cmd CommandOptions) (string, error) {
	base := kdbx.Base(entry)
	dir := kdbx.Directory(entry)
	existing, err := cmd.Transaction().Get(dir, kdbx.SecretValue)
	if err != nil {
		return "", err
	}
	if existing == nil {
		return "", errors.New("entry does not exist")
	}
	val, ok := existing.Value(base)
	if !ok {
		return "", fmt.Errorf("entity value invalid: %s (%s)", base, entry)
	}
	return val, nil
}
