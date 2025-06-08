// Package app can list entries
package app

import (
	"errors"
	"fmt"
	"regexp"
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
	var re *regexp.Regexp
	hasFilter := filter != ""
	if hasFilter {
		var err error
		re, err = regexp.Compile(filter)
		if err != nil {
			return err
		}
	}
	opts := kdbx.QueryOptions{}
	opts.Mode = kdbx.ListMode
	e, err := cmd.Transaction().QueryCallback(opts)
	if err != nil {
		return err
	}
	allowed := func(p string) bool {
		if hasFilter {
			return re.MatchString(p)
		}
		return true
	}
	w := cmd.Writer()
	attrFilter := attr != ""
	for f, err := range e {
		if err != nil {
			return err
		}
		if groups {
			if allowed(f.Path) {
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
			if allowed(path) {
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
