package main

import (
	"fmt"
	"os"

	"github.com/enckse/lockbox/internal/encrypt"
	"github.com/enckse/lockbox/internal/misc"
)

func main() {
	args := os.Args
	result, err := encrypt.FromFile(args[len(args)-1])
	if err != nil {
		misc.Die("unable to read file", err)
	}
	if result != nil {
		fmt.Println(string(result))
	}
}
