// Package app can do various conversions
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/kdbx"
)

// Conv will convert 1-N files
func Conv(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) == 0 {
		return errors.New("conv requires a file")
	}
	w := cmd.Writer()
	for _, a := range args {
		t, err := kdbx.Load(a)
		if err != nil {
			return err
		}
		if err := serialize(w, t, false, ""); err != nil {
			return err
		}
	}
	return nil
}

func serialize(w io.Writer, tx *kdbx.Transaction, isJSON bool, filter string) error {
	e, err := tx.QueryCallback(kdbx.QueryOptions{Mode: kdbx.ListMode, Values: kdbx.JSONValue, PathFilter: filter})
	if err != nil {
		return err
	}
	if isJSON {
		fmt.Fprint(w, "{")
	}
	printed := false
	for item, err := range e {
		if err != nil {
			return err
		}
		if printed {
			if isJSON {
				fmt.Fprint(w, ",")
			}
		}
		if isJSON {
			fmt.Fprint(w, "\n")
		}
		b, err := json.MarshalIndent(map[string]kdbx.EntityValues{item.Path: item.Values}, "", "  ")
		if err != nil {
			return err
		}
		trimmed := strings.TrimSpace(string(b))
		trimmed = strings.TrimPrefix(trimmed, "{")
		trimmed = strings.TrimSuffix(trimmed, "}")
		if isJSON {
			fmt.Fprintf(w, "  %s", strings.TrimSpace(trimmed))
		} else {
			for _, line := range strings.Split(trimmed, "\n") {
				if strings.TrimSpace(line) == "" {
					continue
				}
				fmt.Fprintf(w, "%s\n", strings.TrimPrefix(line, "  "))
			}
		}
		printed = true
	}
	if isJSON {
		if printed {
			fmt.Fprint(w, "\n")
		}
		fmt.Fprint(w, "}\n")
	}
	return nil
}
