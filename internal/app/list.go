// Package app can list entries
package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/backend"
)

// List will list/find entries
func List(cmd CommandOptions, isFind, groups bool) error {
	if isFind && groups {
		return errors.New("groups+find not supported")
	}
	args := cmd.Args()
	filter := ""
	if isFind {
		if len(args) != 1 {
			return errors.New("find requires one argument")
		}
		filter = args[0]
	} else {
		if len(args) != 0 {
			return errors.New("arguments not supported")
		}
	}

	return doList("", filter, cmd, groups)
}

func doList(attr, filter string, cmd CommandOptions, groups bool) error {
	opts := backend.QueryOptions{}
	opts.Mode = backend.ListMode
	opts.PathFilter = filter
	e, err := cmd.Transaction().QueryCallback(opts)
	if err != nil {
		return err
	}
	w := cmd.Writer()
	attrFilter := attr != ""
	for f, err := range e {
		if err != nil {
			return err
		}
		if groups {
			fmt.Fprintf(w, "%s\n", f.Path)
			continue
		}
		if f.Values == nil {
			continue
		}
		var results []string
		for k := range f.Values {
			if attrFilter {
				if k != attr {
					continue
				}
			}
			results = append(results, backend.NewPath(f.Path, k))
		}
		if len(results) == 0 {
			continue
		}
		sort.Strings(results)
		fmt.Fprintf(w, "%s\n", strings.Join(results, "\n"))
	}
	return nil
}
