// Package subcommands handles git diffs.
package subcommands

import (
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/backend"
)

// GitDiff handles git diffing of lb entries.
func GitDiff(args []string) error {
	if len(args) == 0 {
		return errors.New("git diff requires a file")
	}
	t, err := backend.Load(args[len(args)-1])
	if err != nil {
		return err
	}
	e, err := t.QueryCallback(backend.QueryOptions{Mode: backend.ListMode, Values: backend.HashedValue})
	if err != nil {
		return err
	}
	for _, item := range e {
		fmt.Printf("%s:\nhash:%s\n", item.Path, item.Value)
	}
	return nil
}
