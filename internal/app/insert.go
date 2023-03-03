// Package app can insert
package app

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/totp"
)

func insertError(message string, err error) error {
	return fmt.Errorf("%s (%w)", message, err)
}

// Insert will insert new entries
// NOTE: almost entirely tested via regresssion due to complexities around piping/inputs
func Insert(w io.Writer, t *backend.Transaction, args []string, cmd InsertOptions) error {
	multi := false
	isTOTP := false
	idx := 0
	switch len(args) {
	case 0:
		return errors.New("insert requires an entry")
	case 1:
	case 2:
		opt := args[0]
		switch opt {
		case cli.InsertMultiCommand:
			multi = true
		case cli.InsertTOTPCommand:
			off, err := cmd.IsNoTOTP()
			if err != nil {
				return err
			}
			if off {
				return totp.ErrNoTOTP
			}
			isTOTP = true
		default:
			return errors.New("unknown argument")
		}
		multi = true
		idx = 1
	default:
		return errors.New("too many arguments")
	}
	isPipe := cmd.IsPipe()
	entry := args[idx]
	if isTOTP {
		totpToken := cmd.TOTPToken()
		if !strings.HasSuffix(entry, backend.NewSuffix(totpToken)) {
			entry = backend.NewPath(entry, totpToken)
		}
	}
	existing, err := t.Get(entry, backend.BlankValue)
	if err != nil {
		return insertError("unable to check for existing entry", err)
	}
	if existing != nil {
		if !isPipe {
			if !cmd.Confirm("overwrite existing") {
				return nil
			}
		}
	}
	password, err := cmd.Input(isPipe, multi)
	if err != nil {
		return insertError("invalid input", err)
	}
	p := strings.TrimSpace(string(password))
	if err := t.Insert(entry, p); err != nil {
		return insertError("failed to insert", err)
	}
	if !isPipe {
		fmt.Fprintln(w)
	}
	return nil
}
