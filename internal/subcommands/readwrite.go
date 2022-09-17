// Package subcommands perform a read/write against a specific lockbox object.
package subcommands

import (
	"errors"
	"flag"
	"fmt"

	"github.com/enckse/lockbox/internal/encrypt"
)

// ReadWrite performs singular read/write encryption operations.
func ReadWrite(args []string) error {
	flags := flag.NewFlagSet("readwrite", flag.ExitOnError)
	mode := flags.String("mode", "", "decrypt/encrypt")
	key := flags.String("key", "", "security key")
	file := flags.String("file", "", "file to process")
	keyMode := flags.String("keymode", "", "key lookup mode")
	algo := flags.String("algorithm", "", "algorithm to use")
	if err := flags.Parse(args); err != nil {
		return err
	}

	l, err := encrypt.NewLockbox(encrypt.LockboxOptions{Key: *key, KeyMode: *keyMode, File: *file, Algorithm: *algo})
	if err != nil {
		return err
	}
	switch *mode {
	case "encrypt":
		if err := l.Encrypt(nil); err != nil {
			return err
		}
	case "decrypt":
		results, err := l.Decrypt()
		if err != nil {
			return err
		}
		fmt.Println(string(results))
	default:
		return errors.New("invalid read/write modeE")
	}
	return nil
}
