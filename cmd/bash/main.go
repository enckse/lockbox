package main

import (
	_ "embed"
	"fmt"
)

var (
	//go:embed completions.bash
	completions string
)

func main() {
	fmt.Println(completions)
}
