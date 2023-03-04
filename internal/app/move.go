package app

import (
	"errors"

	"github.com/enckse/lockbox/internal/backend"
)

// Move is the CLI command to move entries
func Move(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) != 2 {
		return errors.New("src/dst required for move")
	}
	t := cmd.Transaction()
	src := args[0]
	dst := args[1]
	srcExists, err := t.Get(src, backend.SecretValue)
	if err != nil {
		return errors.New("unable to get source entry")
	}
	if srcExists == nil {
		return errors.New("no source object found")
	}
	dstExists, err := t.Get(dst, backend.BlankValue)
	if err != nil {
		return errors.New("unable to get destination object")
	}
	if dstExists != nil {
		if !cmd.Confirm("overwrite destination") {
			return nil
		}
	}
	return t.Move(*srcExists, dst)
}
