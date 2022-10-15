// provides the binary runs or calls lockbox commands.
package main

import (
	"fmt"
	"os"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/program"
)

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	p, err := program.NewProgram(inputs.ConfirmYesNoPrompt, exit)
	if err != nil {
		exit(err)
	}
	if err := p.Run(os.Args); err != nil {
		exit(err)
	}
}