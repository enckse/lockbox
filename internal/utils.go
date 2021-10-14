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

func isYesNoEnv(defaultValue bool, env string) (bool, error) {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(env)))
	if len(value) == 0 {
		return defaultValue, nil
	}
	switch value {
	case "no":
		return false, nil
	case "yes":
		return true, nil
	}
	return false, stock.NewBasicError(fmt.Sprintf("invalid yes/no env value for %s", env))
}

// IsInteractive indicates if running as a user UI experience.
func IsInteractive() (bool, error) {
	return isYesNoEnv(true, "LOCKBOX_INTERACTIVE")
}

// GetColor will retrieve start/end terminal coloration indicators.
func GetColor(color Color) (string, string, error) {
	if color != ColorRed {
		return "", "", stock.NewBasicError("bad color")
	}
	interactive, err := IsInteractive()
	if err != nil {
		return "", "", err
	}
	colors := interactive
	if colors {
		isColored, err := isYesNoEnv(false, "LOCKBOX_NOCOLOR")
		if err != nil {
			return "", "", err
		}
		colors = isColored
	}
	if colors {
		return termBeginRed, termEndRed, nil
	}
	return "", "", nil
}

// GetStore gets the lockbox directory.
func GetStore() string {
	return os.Getenv("LOCKBOX_STORE")
}

// Find will find all lockbox files in a directory store.
func Find(store string, display bool) ([]string, error) {
	var results []string
	if !stock.PathExists(store) {
		return nil, stock.NewBasicError("store does not exists")
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
		return "", stock.NewBasicError("passwords do NOT match")
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

// IsInputFromPipe will indicate if connected to stdin pipe.
func IsInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
