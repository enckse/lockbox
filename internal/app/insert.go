// Package app can insert
package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/totp"
)

type (
	// InsertOptions are functions required for insert
	InsertOptions interface {
		CommandOptions
		IsPipe() bool
		Input(bool, bool) ([]byte, error)
		TOTPToken() string
		IsNoTOTP() (bool, error)
	}
	// InsertArgs are parsed insert settings for insert commands
	InsertArgs struct {
		Entry string
		Multi bool
	}
)

// ReadArgs will read and check insert args
func ReadArgs(cmd InsertOptions) (InsertArgs, error) {
	multi := false
	isTOTP := false
	idx := 0
	noTOTP, err := cmd.IsNoTOTP()
	if err != nil {
		return InsertArgs{}, err
	}
	args := cmd.Args()
	switch len(args) {
	case 0:
		return InsertArgs{}, errors.New("insert requires an entry")
	case 1:
	case 2:
		opt := args[0]
		switch opt {
		case cli.InsertMultiCommand:
			multi = true
		case cli.InsertTOTPCommand:
			if noTOTP {
				return InsertArgs{}, totp.ErrNoTOTP
			}
			isTOTP = true
		default:
			return InsertArgs{}, errors.New("unknown argument")
		}
		idx = 1
	default:
		return InsertArgs{}, errors.New("too many arguments")
	}
	entry := args[idx]
	if !noTOTP {
		totpToken := cmd.TOTPToken()
		hasSuffixTOTP := strings.HasSuffix(entry, backend.NewSuffix(totpToken))
		if isTOTP {
			if !hasSuffixTOTP {
				entry = backend.NewPath(entry, totpToken)
			}
		} else {
			if hasSuffixTOTP {
				return InsertArgs{}, errors.New("can not insert totp entry without totp flag")
			}
		}

	}
	return InsertArgs{Multi: multi, Entry: entry}, nil
}

// Do will execute an insert
func (args InsertArgs) Do(cmd InsertOptions) error {
	t := cmd.Transaction()
	existing, err := t.Get(args.Entry, backend.BlankValue)
	if err != nil {
		return err
	}
	isPipe := cmd.IsPipe()
	if existing != nil {
		if !isPipe {
			if !cmd.Confirm("overwrite existing") {
				return nil
			}
		}
	}
	password, err := cmd.Input(isPipe, args.Multi)
	if err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	p := strings.TrimSpace(string(password))
	if err := t.Insert(args.Entry, p); err != nil {
		return err
	}
	if !isPipe {
		fmt.Fprintln(cmd.Writer())
	}
	return nil
}
