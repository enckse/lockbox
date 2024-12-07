// Package app handles rekeying a database
package app

import (
	"errors"
	"flag"
	"strings"
)

var reKeyFlags = struct {
	KeyFile string
	NoKey   string
}{"keyfile", "nokey"}

type reKeyArgs struct {
	NoKey   bool
	KeyFile string
}

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
		p, err := cmd.Input(!piping)
		if err != nil {
			return err
		}
		pass = string(p)
	}
	return cmd.Transaction().ReKey(pass, vars.KeyFile)
}

func readArgs(args []string) (reKeyArgs, error) {
	set := flag.NewFlagSet("rekey", flag.ExitOnError)
	keyFile := set.String(reKeyFlags.KeyFile, "", "new keyfile")
	noKey := set.Bool(reKeyFlags.NoKey, false, "disable password/key credential")
	if err := set.Parse(args); err != nil {
		return reKeyArgs{}, err
	}
	noPass := *noKey
	file := *keyFile
	if strings.TrimSpace(file) == "" && noPass {
		return reKeyArgs{}, errors.New("a key or keyfile must be passed for rekey")
	}
	return reKeyArgs{KeyFile: file, NoKey: noPass}, nil
}
