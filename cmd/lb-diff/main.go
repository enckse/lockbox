package main

import (
	"fmt"
	"os"

	"voidedtech.com/lockbox/internal"
	"voidedtech.com/stock"
)

func main() {
	args := os.Args
	l, err := internal.NewLockbox("", "", args[len(args)-1])
	if err != nil {
		stock.Die("unable to make lockbox model instance", err)
	}
	result, err := l.Decrypt()
	if err != nil {
		stock.Die("unable to read file", err)
	}
	if result != nil {
		fmt.Println(string(result))
	}
}
