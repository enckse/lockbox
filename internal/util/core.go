// Package util provides some common operations
package util

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// PathExists will indicate if a path exists
func PathExists(file string) bool {
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// Fatal will call Die but without a message
func Fatal(err error) {
	Die("", err)
}

// Die will write to stderr and exit (1)
func Die(message string, err error) {
	msg := message
	if err != nil {
		if msg == "" {
			msg = err.Error()
		} else {
			msg = fmt.Sprintf("%s (%v)", msg, err)
		}
	}
	if msg != "" {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
	os.Exit(1)
}

// Copy will copy a file from source to destination via ReadFile/WriteFile
func Copy(src, dst string, mode fs.FileMode) error {
	if !PathExists(src) {
		return fmt.Errorf("source file '%s' does not exist", src)
	}

	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.WriteFile(dst, in, mode); err != nil {
		return err
	}

	return nil
}

// ReadStdin will read one (or more) stdin lines
func ReadStdin(one bool) ([]byte, error) {
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
