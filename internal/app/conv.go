package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

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
		if err := serialize(w, t); err != nil {
			return err
		}
	}
	return nil
}

func serialize(w io.Writer, tx *backend.Transaction) error {
	e, err := tx.QueryCallback(backend.QueryOptions{Mode: backend.ListMode, Values: backend.JSONValue})
	if err != nil {
		return err
	}
	fmt.Fprint(w, "{\n")
	for idx, item := range e {
		if idx > 0 {
			fmt.Fprintf(w, ",\n")
		}
		var buf bytes.Buffer
		if err := json.Indent(&buf, []byte(item.Value), "  ", "  "); err != nil {
			return err
		}
		fmt.Fprintf(w, "  \"%s\": %s\n", item.Path, buf.String())
	}
	fmt.Fprintf(w, "}")
	return nil
}
