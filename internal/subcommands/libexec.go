// Package subcommands handles calling library commands.
package subcommands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/enckse/lockbox/internal/misc"
)

type (
	// LibExecOptions for calling to subcommands that are executables.
	LibExecOptions struct {
		Directory string
		Command   string
		Args      []string
	}
)

// LibExecCallback will handle subcommand handling outside of known functions.
func LibExecCallback(args LibExecOptions) error {
	tryCommand := fmt.Sprintf(filepath.Join(args.Directory, "lb-%s"), args.Command)
	if !misc.PathExists(tryCommand) {
		return errors.New(args.Command)
	}
	c := exec.Command(tryCommand, args.Args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
