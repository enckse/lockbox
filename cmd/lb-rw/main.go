// perform a read/write against a specific lockbox object.
package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/misc"
)

func main() {
	mode := flag.String("mode", "", "decrypt/encrypt")
	key := flag.String("key", "", "security key")
	file := flag.String("file", "", "file to process")
	keyMode := flag.String("keymode", "", "key lookup mode")
	flag.Parse()
	l, err := encrypt.NewLockbox(encrypt.LockboxOptions{Key: *key, KeyMode: *keyMode, File: *file})
	if err != nil {
		misc.Die("unable to make lockbox model instance", err)
	}
	switch *mode {
	case "encrypt":
		if err := l.Encrypt(nil); err != nil {
			misc.Die("failed to encrypt", err)
		}
	case "decrypt":
		results, err := l.Decrypt()
		if err != nil {
			misc.Die("failed to decrypt", err)
		}
		fmt.Println(string(results))
	default:
		misc.Die("invalid mode", errors.New("bad mode"))
	}
}
