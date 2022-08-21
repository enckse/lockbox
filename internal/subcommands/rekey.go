// Package subcommands handles rekeying.
package subcommands

import (
	"flag"
	"fmt"
	"strings"

	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/store"
)

// Rekey handles rekeying a lockbox entirely.
func Rekey(args []string) error {
	flags := flag.NewFlagSet("rekey", flag.ExitOnError)
	inKey := flags.String("inkey", "", "input encryption key to read current values")
	outKey := flags.String("outkey", "", "output encryption key to update values with")
	inMode := flags.String("inmode", "", "input encryption key mode")
	outMode := flags.String("outmode", "", "output encryption key mode")
	if err := flags.Parse(args); err != nil {
		return err
	}
	found, err := store.NewFileSystemStore().List(store.ViewOptions{})
	if err != nil {
		return err
	}
	inOpts := encrypt.LockboxOptions{Key: *inKey, KeyMode: *inMode}
	outOpts := encrypt.LockboxOptions{Key: *outKey, KeyMode: *outMode}
	for _, file := range found {
		fmt.Printf("rekeying: %s\n", file)
		inOpts.File = file
		in, err := encrypt.NewLockbox(inOpts)
		if err != nil {
			return err
		}
		decrypt, err := in.Decrypt()
		if err != nil {
			return err
		}
		outOpts.File = file
		out, err := encrypt.NewLockbox(outOpts)
		if err != nil {
			return err
		}
		if err := out.Encrypt([]byte(strings.TrimSpace(string(decrypt)))); err != nil {
			return err
		}
	}
	return nil
}
