package util

import (
	"path/filepath"
)

// ReadDirFile will read a dir+file
func ReadDirFile(dir, file string, e interface{ ReadFile(string) ([]byte, error) }) (string, error) {
	b, err := e.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return "", err
	}
	return string(b), err
}
