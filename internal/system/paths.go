// Package system is responsible for pathing operations/commands
package system

import (
	"errors"
	"os"
)

// PathExists indicates whether a path exists (true) or not (false)
func PathExists(file string) bool {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
