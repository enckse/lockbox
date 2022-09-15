// Package store handles filesystem operations for a lockbox store.
package store

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
)

const (
	extension = ".lb"
)

type (
	// ListEntryFilter allows for filtering/changing view results.
	ListEntryFilter func(string) string
	// FileSystem represents a filesystem store.
	FileSystem struct {
		path string
	}
	// ViewOptions represent list options for parsing store entries.
	ViewOptions struct {
		Display      bool
		Filter       ListEntryFilter
		ErrorOnEmpty bool
	}
)

// NewFileSystemStore gets the lockbox directory (filesystem-based) store.
func NewFileSystemStore() FileSystem {
	return FileSystem{path: os.Getenv(inputs.StoreEnv)}
}

// List will get all lockbox files in a store.
func (s FileSystem) List(options ViewOptions) ([]string, error) {
	var results []string
	if !PathExists(s.path) {
		return nil, errors.New("store does not exist")
	}
	err := filepath.Walk(s.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, extension) {
			usePath := path
			if options.Display {
				usePath = s.trim(usePath)
			}
			if options.Filter != nil {
				usePath = options.Filter(usePath)
				if usePath == "" {
					return nil
				}
			}
			results = append(results, usePath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if options.ErrorOnEmpty && len(results) == 0 {
		return nil, errors.New("no results found")
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
	if !strings.HasSuffix(file, extension) {
		return file + extension
	}
	return file
}

// CleanPath will clean store and extension information from an entry.
func (s FileSystem) CleanPath(fullPath string) string {
	return s.trim(fullPath)
}

func (s FileSystem) trim(path string) string {
	f := strings.TrimPrefix(path, s.path)
	f = strings.TrimPrefix(f, string(os.PathSeparator))
	return strings.TrimSuffix(f, extension)
}

// PathExists indicates if a path exists.
func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GitCommit is for adding/changing entities
func (s FileSystem) GitCommit(entry string) error {
	return s.gitAction("add", entry)
}

// GitRemove is for removing entities
func (s FileSystem) GitRemove(entry string) error {
	return s.gitAction("rm", entry)
}

func (s FileSystem) gitAction(action, entry string) error {
	ok, err := inputs.IsGitEnabled()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if !PathExists(filepath.Join(s.path, ".git")) {
		return nil
	}
	useEntry, err := filepath.Rel(s.path, entry)
	if err != nil {
		return err
	}
	if err := s.gitRun(action, useEntry); err != nil {
		return err
	}
	return s.gitRun("commit", "-m", fmt.Sprintf("lb %s: %s", action, useEntry))
}

func (s FileSystem) gitRun(args ...string) error {
	arguments := []string{"-C", s.path}
	arguments = append(arguments, args...)
	cmd := exec.Command("git", arguments...)
	ok, err := inputs.IsGitQuiet()
	if err != nil {
		return err
	}
	if !ok {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}
