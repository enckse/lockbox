package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"voidedtech.com/stock"
)

type (
	// Color are terminal colors for dumb terminal coloring.
	Color int
)

const (
	// Extension is the lockbox file extension.
	Extension    = ".lb"
	termBeginRed = "\033[1;31m"
	termEndRed   = "\033[0m"
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

func termEcho(on bool) {
	// Common settings and variables for both stty calls.
	attrs := syscall.ProcAttr{
		Dir:   "",
		Env:   []string{},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Sys:   nil}
	var ws syscall.WaitStatus
	cmd := "echo"
	if !on {
		cmd = "-echo"
	}

	// Enable/disable echoing.
	pid, err := syscall.ForkExec(
		"/bin/stty",
		[]string{"stty", cmd},
		&attrs)
	if err != nil {
		panic(err)
	}

	// Wait for the stty process to complete.
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		panic(err)
	}
}

// ConfirmInput will get 2 inputs and confirm they are the same.
func ConfirmInput() (string, error) {
	termEcho(false)
	defer func() {
		termEcho(true)
	}()
	fmt.Printf("please enter password: ")
	first, err := Stdin(true)
	if err != nil {
		return "", err
	}
	fmt.Printf("\nplease re-enter password: ")
	second, err := Stdin(true)
	if err != nil {
		return "", err
	}
	if first != second {
		return "", NewLockboxError("passwords do NOT match")
	}
	return first, nil
}

// Stdin will retrieve stdin data.
func Stdin(one bool) (string, error) {
	b, err := stock.Stdin(one)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
