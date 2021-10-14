package main

import (
	"flag"
	"fmt"

	"voidedtech.com/lockbox/internal"
	"voidedtech.com/stock"
)

func main() {
	mode := flag.String("mode", "", "decrypt/encrypt")
	key := flag.String("key", "", "security key")
	file := flag.String("file", "", "file to process")
	keyMode := flag.String("keymode", "", "key lookup mode")
	flag.Parse()
	l, err := internal.NewLockbox(*key, *keyMode, *file)
	if err != nil {
		stock.Die("unable to make lockbox model instance", err)
	}
	switch *mode {
	case "encrypt":
		if err := l.Encrypt(nil); err != nil {
			stock.Die("failed to encrypt", err)
		}
	case "decrypt":
		results, err := l.Decrypt()
		if err != nil {
			stock.Die("failed to decrypt", err)
		}
		fmt.Println(string(results))
	default:
		stock.Die("invalid mode", stock.NewBasicError("bad mode"))
	}
}
