package main

import (
	"flag"
	"fmt"
	"strings"

	"voidedtech.com/lockbox/internal"
)

func main() {
	inKey := flag.String("inkey", "", "input encryption key to read current values")
	outKey := flag.String("outkey", "", "output encryption key to update values with")
	inMode := flag.String("inmode", "", "input encryption key mode")
	outMode := flag.String("outmode", "", "output encryption key mode")
	flag.Parse()
	found, err := internal.Find(internal.GetStore(), false)
	if err != nil {
		internal.Die("failed finding entries", err)
	}
	for _, file := range found {
		fmt.Printf("rekeying: %s\n", file)
		in, err := internal.NewLockbox(*inKey, *inMode, file)
		if err != nil {
			internal.Die("unable to make input lockbox", err)
		}
		decrypt, err := in.Decrypt()
		if err != nil {
			internal.Die("failed to process file decryption", err)
		}
		out, err := internal.NewLockbox(*outKey, *outMode, file)
		if err != nil {
			internal.Die("unable to make output lockbox", err)
		}
		if err := out.Encrypt([]byte(strings.TrimSpace(string(decrypt)))); err != nil {
			internal.Die("failed to encrypt file", err)
		}
	}
}
