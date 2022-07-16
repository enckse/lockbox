package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"github.com/enckse/lockbox/internal/misc"
)

const (
	// Extension is the lockbox file extension.
	Extension = ".lb"
)

type (
	// FileSystem represents a filesystem store.
	FileSystem struct {
		path string
	}
	ViewOptions struct {
		Display bool
	}

)

// NewFileSystemStore gets the lockbox directory (filesystem-based) store.
func NewFileSystemStore() string {
	return os.Getenv("LOCKBOX_STORE")
}

// List will get all lockbox files in a store.
func (s FileSystem) List(options ViewOptions) ([]string, error) {
	var results []string
	if !misc.PathExists(s.path) {
		return nil, errors.New("store does not exist")
	}
	err := filepath.Walk(s.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, Extension) {
			usePath := path
			if options.Display {
				usePath = strings.TrimPrefix(usePath, s.path)
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
	if options.Display {
		sort.Strings(results)
	}
	return results, nil
}
