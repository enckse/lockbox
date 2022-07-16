package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/misc"
	"github.com/enckse/lockbox/internal/store"
)

func main() {
	inKey := flag.String("inkey", "", "input encryption key to read current values")
	outKey := flag.String("outkey", "", "output encryption key to update values with")
	inMode := flag.String("inmode", "", "input encryption key mode")
	outMode := flag.String("outmode", "", "output encryption key mode")
	flag.Parse()
	found, err := store.NewFileSystemStore().List(store.ViewOptions{})
	if err != nil {
		misc.Die("failed finding entries", err)
	}
	inOpts := encrypt.LockboxOptions{Key: *inKey, KeyMode: *inMode}
	outOpts := encrypt.LockboxOptions{Key: *outKey, KeyMode: *outMode}
	for _, file := range found {
		fmt.Printf("rekeying: %s\n", file)
		inOpts.File = file
		in, err := encrypt.NewLockbox(inOpts)
		if err != nil {
			misc.Die("unable to make input lockbox", err)
		}
		decrypt, err := in.Decrypt()
		if err != nil {
			misc.Die("failed to process file decryption", err)
		}
		outOpts.File = file
		out, err := encrypt.NewLockbox(outOpts)
		if err != nil {
			misc.Die("unable to make output lockbox", err)
		}
		if err := out.Encrypt([]byte(strings.TrimSpace(string(decrypt)))); err != nil {
			misc.Die("failed to encrypt file", err)
		}
	}
}
