// Package app can insert
package app

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
)

func insertError(message string, err error) error {
	return fmt.Errorf("%s (%w)", message, err)
}

// ParseInsertArgs will parse the input args for insert commands
func ParseInsertArgs(cmd InsertOptions, args []string) (InsertArgs, error) {
	multi := false
	idx := 0
	switch len(args) {
	case 0:
		return InsertArgs{}, errors.New("insert requires an entry")
	case 1:
	case 2:
		opt := args[0]
		switch opt {
		case cli.InsertMultiCommand:
			multi = true
		default:
			return InsertArgs{}, errors.New("unknown argument")
		}
		multi = true
		idx = 1
	default:
		return InsertArgs{}, errors.New("too many arguments")
	}
	return InsertArgs{Opts: cmd, Multi: multi, Entry: args[idx]}, nil
}

// Do will execute an insert
func (args InsertArgs) Do(w io.Writer, t *backend.Transaction) error {
	existing, err := t.Get(args.Entry, backend.BlankValue)
	if err != nil {
		return insertError("unable to check for existing entry", err)
	}
	isPipe := args.Opts.IsPipe()
	if existing != nil {
		if !isPipe {
			if !args.Opts.Confirm("overwrite existing") {
				return nil
			}
		}
	}
	password, err := args.Opts.Input(isPipe, args.Multi)
	if err != nil {
		return insertError("invalid input", err)
	}
	p := strings.TrimSpace(string(password))
	if err := t.Insert(args.Entry, p); err != nil {
		return insertError("failed to insert", err)
	}
	if !isPipe {
		fmt.Fprintln(w)
	}
	return nil
}
