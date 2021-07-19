package main

import (
	"flag"
	"fmt"

	"voidedtech.com/lockbox/internal"
)

func main() {
	mode := flag.String("mode", "", "decrypt/encrypt")
	key := flag.String("key", "", "security key")
	file := flag.String("file", "", "file to process")
	keyMode := flag.String("keymode", "", "key lookup mode")
	flag.Parse()
	l, err := internal.NewLockbox(*key, *keyMode, *file)
	if err != nil {
		internal.Die("unable to make lockbox model instance", err)
	}
	switch *mode {
	case "encrypt":
		if err := l.Encrypt(nil); err != nil {
			internal.Die("failed to encrypt", err)
		}
	case "decrypt":
		results, err := l.Decrypt()
		if err != nil {
			internal.Die("failed to decrypt", err)
		}
		fmt.Println(string(results))
	default:
		internal.Die("invalid mode", fmt.Errorf("bad mode"))
	}
}
