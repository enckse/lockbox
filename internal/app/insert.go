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

type (
	// InsertOptions are functions required for insert
	InsertOptions struct {
		IsPipe  func() bool
		Input   func(bool, bool) ([]byte, error)
		Confirm Confirm
	}
	// InsertArgsOptions supports cli arg parsing
	InsertArgsOptions struct {
		TOTPToken func() string
		IsNoTOTP  func() (bool, error)
	}
	// InsertArgs are parsed insert settings for insert commands
	InsertArgs struct {
		Entry string
		Multi bool
		Opts  InsertOptions
	}
)

func insertError(message string, err error) error {
	return fmt.Errorf("%s (%w)", message, err)
}

// ReadArgs will read and check insert args
func (p InsertArgsOptions) ReadArgs(cmd InsertOptions, args []string) (InsertArgs, error) {
	multi := false
	isTOTP := false
	idx := 0
	noTOTP, err := p.IsNoTOTP()
	if err != nil {
		return InsertArgs{}, err
	}
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
		totpToken := p.TOTPToken()
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
	return InsertArgs{Opts: cmd, Multi: multi, Entry: entry}, nil
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
