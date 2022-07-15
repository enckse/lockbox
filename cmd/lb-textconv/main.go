package main

import (
	"fmt"
	"os"

	"github.com/enckse/lockbox/internal"
)

func main() {
	args := os.Args
	if len(args) != 3 {
		internal.Die("input entry required", internal.NewLockboxError("entry argument required"))
	}
	l, err := internal.NewLockbox(internal.LockboxOptions{File: args[len(args)-1]})
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
