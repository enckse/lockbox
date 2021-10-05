package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"voidedtech.com/stock"
)

type (
	// Color are terminal colors for dumb terminal coloring.
	Color int
)

const (
	// Extension is the lockbox file extension.
	Extension = ".lb"
	termBeginRed = "\033[1;31m"
	termEndRed = "\033[0m"
	// ColorRed will get red terminal coloring.
	ColorRed = iota
)

// GetColor will retrieve start/end terminal coloration indicators.
func GetColor(color Color) (string, string, error) {
	if color != ColorRed {
		return "", "", NewLockboxError("bad color")
	}
	if os.Getenv("LOCKBOX_NOCOLOR") == "yes" {
		return "", "", nil
	}
	return termBeginRed, termEndRed, nil
}

// GetStore gets the lockbox directory.
func GetStore() string {
	return os.Getenv("LOCKBOX_STORE")
}

// Find will find all lockbox files in a directory store.
func Find(store string, display bool) ([]string, error) {
	var results []string
	if !stock.PathExists(store) {
		return nil, NewLockboxError("store does not exists")
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
