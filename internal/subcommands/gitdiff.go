// Package subcommands handles git diffs.
package subcommands

import (
	"errors"
	"fmt"

	"github.com/enckse/lockbox/internal/encrypt"
)

// GitDiff handles git diffing of lb entries.
func GitDiff(args []string) error {
	if len(args) == 0 {
		return errors.New("git diff requires a file")
	}
	result, err := encrypt.FromFile(args[len(args)-1])
	if err != nil {
		return err
	}
	if result != nil {
		fmt.Println(string(result))
	}
	return nil
}
