// Package app handles rekeying a database
package app

import (
	"errors"
	"flag"
	"strings"

	"github.com/enckse/lockbox/internal/app/commands"
)

// ReKey handles entry rekeying
func ReKey(cmd UserInputOptions) error {
	args := cmd.Args()
	vars, err := readArgs(args)
	if err != nil {
		return err
	}
	piping := cmd.IsPipe()
	if !piping {
		if !cmd.Confirm("proceed with rekey") {
			return nil
		}
	}
	var pass string
	if !vars.NoKey {
		p, err := cmd.Input(!piping, true, "password")
		if err != nil {
			return err
		}
		pass = string(p)
	}
	return cmd.Transaction().ReKey(pass, vars.KeyFile)
}

func readArgs(args []string) (commands.ReKeyArgs, error) {
	set := flag.NewFlagSet("rekey", flag.ExitOnError)
	keyFile := set.String(commands.ReKeyFlags.KeyFile, "", "new keyfile")
	noKey := set.Bool(commands.ReKeyFlags.NoKey, false, "disable password/key credential")
	if err := set.Parse(args); err != nil {
		return commands.ReKeyArgs{}, err
	}
	noPass := *noKey
	file := *keyFile
	if strings.TrimSpace(file) == "" && noPass {
		return commands.ReKeyArgs{}, errors.New("a key or keyfile must be passed for rekey")
	}
	return commands.ReKeyArgs{KeyFile: file, NoKey: noPass}, nil
}
