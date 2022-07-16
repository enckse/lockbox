package inputs

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"
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
	return false, fmt.Errorf("invalid yes/no env value for %s", env)
}

func IsColorEnabled() (bool, error) {
	return isYesNoEnv(false, "LOCKBOX_NOCOLOR")
}

// IsInteractive indicates if running as a user UI experience.
func IsInteractive() (bool, error) {
	return isYesNoEnv(true, "LOCKBOX_INTERACTIVE")
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

// ConfirmInputsMatch will get 2 inputs and confirm they are the same.
func ConfirmInputsMatch(object string) (string, error) {
	termEcho(false)
	defer func() {
		termEcho(true)
	}()
	fmt.Printf("please enter %s: ", object)
	first, err := Stdin(true)
	if err != nil {
		return "", err
	}
	fmt.Printf("\nplease re-enter %s: ", object)
	second, err := Stdin(true)
	if err != nil {
		return "", err
	}
	if first != second {
		return "", fmt.Errorf("%s(s) do NOT match", object)
	}
	return first, nil
}

// Stdin will retrieve stdin data.
func Stdin(one bool) (string, error) {
	b, err := getStdin(one)
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

// RawStdin will get raw stdin data.
func RawStdin() ([]byte, error) {
	return getStdin(false)
}

func getStdin(one bool) ([]byte, error) {
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
