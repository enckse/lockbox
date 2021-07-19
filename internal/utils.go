package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// Extension is the lockbox file extension
	Extension = ".lb"
)

// GetStore gets the lockbox directory
func GetStore() string {
	return os.Getenv("LOCKBOX_STORE")
}

// Find will find all lockbox files in a directory store
func Find(store string, display bool) ([]string, error) {
	var results []string
	if !PathExists(store) {
		return nil, fmt.Errorf("store does not exists")
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

// Die will print messages and exit
func Die(message string, err error) {
	msg := fmt.Sprintf("%s (%v)", message, err)
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// Stdin reads one (or more) lines from stdin
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

// PathExists indicates if a path exists
func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
