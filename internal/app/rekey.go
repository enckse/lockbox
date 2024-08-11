package app

import (
	"errors"
	"fmt"

	"github.com/seanenck/lockbox/internal/config"
)

type (
	// KeyerOptions defines how rekeying happens
	KeyerOptions interface {
		CommandOptions
		IsPipe() bool
		Password() (string, error)
		ReadLine() (string, error)
	}
)

func getNewPassword(pipe bool, against string, r KeyerOptions) (string, error) {
	if pipe {
		val, err := r.ReadLine()
		if err != nil {
			return "", err
		}
		return val, nil
	}
	fmt.Print("new ")
	p, err := r.Password()
	if err != nil {
		return "", err
	}
	if against != "" {
		if p != against {
			return "", errors.New("rekey passwords do not match")
		}
	}
	return p, nil
}

// ReKey handles entry rekeying
func ReKey(cmd KeyerOptions) error {
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
		first, err := getNewPassword(piping, "", cmd)
		if err != nil {
			return err
		}
		if !piping {
			fmt.Println()
			if _, err := getNewPassword(piping, first, cmd); err != nil {
				return err
			}
			fmt.Println()
		}
		pass = first
		if pass == "" {
			return errors.New("password required but not given")
		}
	}
	return cmd.Transaction().ReKey(pass, vars.KeyFile)
}
