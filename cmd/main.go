// provides the binary runs or calls lockbox commands.
package main

import (
	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/util"
)

func main() {
	if err := app.Run(); err != nil {
		util.Die(err)
	}
}
