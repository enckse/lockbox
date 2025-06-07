// Package app can list entries
package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/kdbx"
)

// List will list/find entries
func List(cmd CommandOptions, groups bool) error {
	args := cmd.Args()
	filter := ""
	switch len(args) {
	case 0:
		break
	case 1:
		filter = args[0]
	default:
		return errors.New("too many arguments (none or filter)")
	}

	return doList("", filter, cmd, groups)
}

func doList(attr, filter string, cmd CommandOptions, groups bool) error {
	opts := kdbx.QueryOptions{}
	opts.Mode = kdbx.ListMode
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
			results = append(results, kdbx.NewPath(f.Path, k))
		}
		if len(results) == 0 {
			continue
		}
		sort.Strings(results)
		fmt.Fprintf(w, "%s\n", strings.Join(results, "\n"))
	}
	return nil
}
