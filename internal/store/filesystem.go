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
	extension = ".lb"
)

type (
	// FileSystem represents a filesystem store.
	FileSystem struct {
		path string
	}
	// ViewOptions represent list options for parsing store entries.
	ViewOptions struct {
		Display bool
	}
)

// NewFileSystemStore gets the lockbox directory (filesystem-based) store.
func NewFileSystemStore() FileSystem {
	return FileSystem{path: os.Getenv("LOCKBOX_STORE")}
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
		if strings.HasSuffix(path, extension) {
			usePath := path
			if options.Display {
				usePath = strings.TrimPrefix(usePath, s.path)
				usePath = strings.TrimPrefix(usePath, string(os.PathSeparator))
				usePath = strings.TrimSuffix(usePath, extension)
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

// NewPath creates a new filesystem store path for an entry.
func (s FileSystem) NewPath(file string) string {
	return s.NewFile(filepath.Join(s.path, file))
}

// NewFile creates a new file with the proper extension.
func (s FileSystem) NewFile(file string) string {
	return file + extension
}

// CleanPath will clean store and extension information from an entry.
func (s FileSystem) CleanPath(fullPath string) string {
	fileName := fullPath
	if strings.HasPrefix(fullPath, s.path) {
		fileName = fileName[len(s.path):]
	}
	if fileName[0] == os.PathSeparator {
		fileName = fileName[1:]
	}
	if strings.HasSuffix(fileName, extension) {
		fileName = fileName[0 : len(fileName)-len(extension)]
	}
	return fileName
}
