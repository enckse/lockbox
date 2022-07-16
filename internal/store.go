package internal

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// Extension is the lockbox file extension.
	Extension = ".lb"
)

// GetStore gets the lockbox directory.
func GetStore() string {
	return os.Getenv("LOCKBOX_STORE")
}

// List will get all lockbox files in a directory store.
func List(store string, display bool) ([]string, error) {
	var results []string
	if !PathExists(store) {
		return nil, errors.New("store does not exists")
	}
	err := filepath.Walk(store, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, Extension) {
			usePath := path
			if display {
				usePath = strings.TrimPrefix(usePath, store)
				usePath = strings.TrimPrefix(usePath, "/")
				usePath = strings.TrimSuffix(usePath, Extension)
			}
			results = append(results, usePath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if display {
		sort.Strings(results)
	}
	return results, nil
}
