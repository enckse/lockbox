package internal

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"voidedtech.com/stock"
)

const (
	// Extension is the lockbox file extension.
	Extension = ".lb"
)

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

// Stdin reads one (or more) lines from stdin.
func Stdin(one bool) ([]byte, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var b bytes.Buffer
	for scanner.Scan() {
		b.WriteString(scanner.Text())
		b.WriteString("\n")
		if one {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
