// Package system handles stdin processing
package system

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

type (
	stdinReaderFunc func(string) (bool, error)
)

func termEcho(on bool) {
	// Common settings and variables for both stty calls.
	attrs := syscall.ProcAttr{
		Dir:   "",
		Env:   []string{},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		Sys:   nil,
	}
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

// GetUserInputPassword will read the user's input from stdin via multiple means.
func GetUserInputPassword(piping, multiLine bool) ([]byte, error) {
	var password string
	if !multiLine && !piping {
		input, err := confirmInputsMatch()
		if err != nil {
			return nil, err
		}
		password = input
	} else {
		input, err := Stdin(false)
		if err != nil {
			return nil, err
		}
		password = input
	}
	if password == "" {
		return nil, errors.New("password can NOT be empty")
	}
	return []byte(password), nil
}

func confirmInputsMatch() (string, error) {
	termEcho(false)
	defer func() {
		termEcho(true)
	}()
	fmt.Print("please enter password: ")
	first, err := Stdin(true)
	if err != nil {
		return "", err
	}
	fmt.Print("\nplease re-enter password: ")
	second, err := Stdin(true)
	if err != nil {
		return "", err
	}
	if first != second {
		return "", errors.New("passwords do NOT match")
	}
	return first, nil
}

// Stdin will get one (or more) lines of stdin as string.
func Stdin(one bool) (string, error) {
	var b []byte
	var err error
	if one {
		b, err = readLine()
	} else {
		b, err = readAll()
	}
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

// ConfirmYesNoPrompt will ask a yes/no question.
func ConfirmYesNoPrompt(prompt string) (bool, error) {
	fmt.Printf("%s? (y/N) ", prompt)
	resp, err := Stdin(true)
	if err != nil {
		return false, err
	}
	return resp == "Y" || resp == "y", nil
}

func readAll() ([]byte, error) {
	return read(false)
}

func readLine() ([]byte, error) {
	return read(true)
}

// ReadFunc will read stdin and execute the given function
func ReadFunc(reader stdinReaderFunc) error {
	if reader == nil {
		return errors.New("invalid reader, nil")
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		ok, err := reader(scanner.Text())
		if err != nil {
			return err
		}
		if !ok {
			break
		}
	}
	return scanner.Err()
}

func read(one bool) ([]byte, error) {
	var b bytes.Buffer
	err := ReadFunc(func(line string) (bool, error) {
		if _, err := b.WriteString(line); err != nil {
			return false, err
		}
		if _, err := b.WriteString("\n"); err != nil {
			return false, err
		}
		if one {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
