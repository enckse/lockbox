// Package app can insert
package app

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/enckse/lockbox/internal/app/totp"
	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/kdbx"
)

// Insert will execute an insert
func Insert(cmd UserInputOptions) error {
	t := cmd.Transaction()
	args := cmd.Args()
	if len(args) != 1 {
		return errors.New("invalid insert, no entry given")
	}
	entry := args[0]
	base := kdbx.Base(entry)
	if !slices.ContainsFunc(kdbx.AllowedFields, func(v string) bool {
		return base == strings.ToLower(v)
	}) {
		return fmt.Errorf("'%s' is not an allowed field name", base)
	}

	dir := kdbx.Directory(entry)
	existing, err := t.Get(dir, kdbx.SecretValue)
	if err != nil {
		return err
	}
	isPipe := cmd.IsPipe()
	if existing != nil {
		if !isPipe {
			if _, ok := existing.Value(base); ok {
				if !cmd.Confirm("overwrite existing") {
					return nil
				}
			}
		}
	}
	isPass := !strings.EqualFold(base, kdbx.URLField)
	password, err := cmd.Input(!isPipe && !strings.EqualFold(base, kdbx.NotesField), isPass, base)
	if err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}
	if !isPipe {
		if isPass {
			fmt.Fprintln(cmd.Writer())
		}
	}
	vals := make(kdbx.EntityValues)
	if existing != nil {
		vals = existing.Values
	}
	cleaned := strings.TrimSpace(string(password))
	if config.EnvTOTPCheckOnInsert.Get() && strings.EqualFold(base, kdbx.OTPField) {
		generator, err := totp.New(cleaned)
		if err != nil {
			return err
		}
		if _, err := generator.Code(); err != nil {
			return err
		}
	}
	vals[base] = cleaned
	if err := t.Insert(dir, vals); err != nil {
		return err
	}
	return nil
}
