package main

import (
	"fmt"
	"os"

	"voidedtech.com/lockbox/internal"
)

func main() {
	args := os.Args
	l, err := internal.NewLockbox("", "", args[len(args)-1])
	if err != nil {
		internal.Die("unable to make lockbox model instance", err)
	}
	result, err := l.Decrypt()
	if err != nil {
		internal.Die("unable to read file", err)
	}
	if result != nil {
		fmt.Println(string(result))
	}
}
