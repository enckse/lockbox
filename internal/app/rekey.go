package app

import (
	"github.com/seanenck/lockbox/internal/config"
)

// ReKey handles entry rekeying
func ReKey(cmd UserInputOptions) error {
	args := cmd.Args()
	vars, err := config.GetReKey(args)
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
		p, err := cmd.Input(piping, false)
		if err != nil {
			return err
		}
		pass = string(p)
	}
	return cmd.Transaction().ReKey(pass, vars.KeyFile)
}
