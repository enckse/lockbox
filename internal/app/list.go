// Package app can list entries
package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/enckse/lockbox/internal/kdbx"
)

// ListMode indicates how listing will be done/output
type ListMode int

const (
	// ListEntriesMode will list the actual entries
	ListEntriesMode ListMode = iota
	// ListGroupsMode will list groups only (e.g. dirnames of all entries)
	ListGroupsMode
	// ListFieldsMode will list groups only, but with ALL possible/allowed field names
	ListFieldsMode
)

// List will list/find entries
func List(cmd CommandOptions, mode ListMode) error {
	args := cmd.Args()
	var filter string
	switch len(args) {
	case 0:
		break
	case 1:
		filter = args[0]
	default:
		return errors.New("too many arguments (none or filter)")
	}

	return doList("", filter, cmd, mode)
}

func doList(attr, filter string, cmd CommandOptions, mode ListMode) error {
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
	isFields := mode == ListFieldsMode
	var allowedFields []string
	if isFields {
		allowedFields = kdbx.AllFieldsLower
	}
	isGroups := mode == ListGroupsMode || isFields
	for f, err := range e {
		if err != nil {
			return err
		}
		if isGroups {
			ok, err := allowed(f.Path)
			if err != nil {
				return err
			}
			if ok {
				output := []string{f.Path}
				if isFields {
					output = []string{}
					for _, allowed := range allowedFields {
						output = append(output, kdbx.NewPath(f.Path, allowed))
					}
				}
				for _, out := range output {
					fmt.Fprintf(w, "%s\n", out)
				}
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
	return true, func(criteria, path string) (bool, error) {
		ok, err := kdbx.Glob(criteria, path)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		return strings.Contains(path, criteria), nil
	}
}
