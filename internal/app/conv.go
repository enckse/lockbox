package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/enckse/lockbox/internal/backend"
)

// Conv will convert 1-N files
func Conv(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) == 0 {
		return errors.New("conv requires a file")
	}
	w := cmd.Writer()
	for _, a := range args {
		t, err := backend.Load(a)
		if err != nil {
			return err
		}
		if err := serialize(w, t, ""); err != nil {
			return err
		}
	}
	return nil
}

func serialize(w io.Writer, tx *backend.Transaction, filter string) error {
	e, err := tx.QueryCallback(backend.QueryOptions{Mode: backend.ListMode, Values: backend.JSONValue})
	if err != nil {
		return err
	}
	fmt.Fprint(w, "{")
	hasFilter := len(filter) > 0
	printed := false
	for _, item := range e {
		if hasFilter {
			if !strings.HasPrefix(item.Path, filter) {
				continue
			}
		}
		if printed {
			fmt.Fprint(w, ",")
		}
		fmt.Fprint(w, "\n")
		b, err := json.MarshalIndent(map[string]json.RawMessage{item.Path: json.RawMessage([]byte(item.Value))}, "", "  ")
		if err != nil {
			return err
		}
		trimmed := strings.TrimSpace(string(b))
		trimmed = strings.TrimPrefix(trimmed, "{")
		trimmed = strings.TrimSuffix(trimmed, "}")
		fmt.Fprintf(w, "  %s", strings.TrimSpace(trimmed))
		printed = true
	}
	if printed {
		fmt.Fprint(w, "\n")
	}
	fmt.Fprint(w, "}\n")
	return nil
}
