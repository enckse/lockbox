// Package app can get stats
package app

import (
	"errors"
)

// JSON will get entries (1 or ALL) in JSON format
func JSON(cmd CommandOptions) error {
	args := cmd.Args()
	if len(args) > 1 {
		return errors.New("invalid arguments")
	}
	filter := ""
	if len(args) == 1 {
		filter = args[0]
	}
	return serialize(cmd.Writer(), cmd.Transaction(), filter)
}
