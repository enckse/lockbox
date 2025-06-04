// Package app can list entries
package app

import (
	"errors"
	"fmt"
	"regexp"

	"git.sr.ht/~enckse/lockbox/internal/backend"
)

// List will list/find entries
func List(cmd CommandOptions, isFind bool) error {
	args := cmd.Args()
	opts := backend.QueryOptions{}
	opts.Mode = backend.ListMode
	if isFind {
		if len(args) != 1 {
			return errors.New("find requires one argument")
		}
	} else {
		if len(args) != 0 {
			return errors.New("list does not support any arguments")
		}
	}
	e, err := cmd.Transaction().QueryCallback(opts)
	if err != nil {
		return err
	}
	w := cmd.Writer()
	printer := func(p string) {
		fmt.Fprintf(w, "%s\n", p)
	}
	finder := printer
	if isFind {
		re, err := regexp.Compile(args[0])
		if err != nil {
			return err
		}
		finder = func(p string) {
			if re.MatchString(p) {
				printer(p)
			}
		}
	}
	for f, err := range e {
		if err != nil {
			return err
		}
		finder(f.Path)
	}
	return nil
}
