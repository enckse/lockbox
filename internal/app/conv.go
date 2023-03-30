package app

import (
	"errors"
	"fmt"
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
		e, err := t.QueryCallback(backend.QueryOptions{Mode: backend.ListMode, Values: backend.HashedValue})
		if err != nil {
			return err
		}
		for _, item := range e {
			fmt.Fprintf(w, "%s:\n  %s\n\n", item.Path, strings.ReplaceAll(item.Value, "\n", "\n  "))
		}
	}
	return nil
}
