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
	hasFilter, selector := createFilter(filter)
	opts := kdbx.QueryOptions{}
	opts.Mode = kdbx.ListMode
	e, err := cmd.Transaction().QueryCallback(opts)
	if err != nil {
		return err
	}
	allowed := func(p string) (bool, error) {
		if hasFilter {
			return selector(filter, p)
		}
		return true, nil
	}
	w := cmd.Writer()
	attrFilter := attr != ""
	for f, err := range e {
		if err != nil {
			return err
		}
		if groups {
			ok, err := allowed(f.Path)
			if err != nil {
				return err
			}
			if ok {
				fmt.Fprintf(w, "%s\n", f.Path)
			}
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
			path := kdbx.NewPath(f.Path, k)
			ok, err := allowed(path)
			if err != nil {
				return err
			}
			if ok {
				results = append(results, path)
			}
		}
		if len(results) == 0 {
			continue
		}
		sort.Strings(results)
		fmt.Fprintf(w, "%s\n", strings.Join(results, "\n"))
	}
	return nil
}

func createFilter(filter string) (bool, func(string, string) (bool, error)) {
	if filter == "" {
		return false, nil
	}
	parts := kdbx.SplitPath(filter)
	for _, p := range parts {
		if strings.ContainsAny(p, "*?[^]\\-") {
			return true, kdbx.Glob
		}
	}
	return true, func(criteria, path string) (bool, error) {
		return strings.Contains(path, criteria), nil
	}
}
