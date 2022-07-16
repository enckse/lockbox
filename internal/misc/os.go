package misc

import (
	"os"
)

// PathExists indicates if a path exists.
func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
