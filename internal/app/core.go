// Package app common objects
package app

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/enckse/lockbox/internal/kdbx"
	"github.com/enckse/lockbox/internal/platform"
)

type (
	// CommandOptions define how commands operate as an application
	CommandOptions interface {
		Confirm(string) bool
		Args() []string
		Transaction() *kdbx.Transaction
		Writer() io.Writer
	}

	// UserInputOptions handle user inputs (e.g. password entry)
	UserInputOptions interface {
		CommandOptions
		IsPipe() bool
		Input(bool, bool, string) ([]byte, error)
	}

	// DefaultCommand is the default CLI app type for actual execution
	DefaultCommand struct {
		tx   *kdbx.Transaction
		args []string
	}

	// ConfigLoader is the default application loader to assist with config handling
	ConfigLoader struct{}
)

// NewDefaultCommand creates a new app command
func NewDefaultCommand(args []string) (*DefaultCommand, error) {
	t, err := kdbx.NewTransaction()
	if err != nil {
		return nil, err
	}
	return &DefaultCommand{args: args, tx: t}, nil
}

// Args will get the args passed to the application
func (a *DefaultCommand) Args() []string {
	return a.args
}

// Writer will get stdout
func (a *DefaultCommand) Writer() io.Writer {
	return os.Stdout
}

// Transaction will return the backend transaction
func (a *DefaultCommand) Transaction() *kdbx.Transaction {
	return a.tx
}

// Confirm will confirm with the user (dying if something abnormal happens)
func (a *DefaultCommand) Confirm(prompt string) bool {
	yesNo, err := platform.ConfirmYesNoPrompt(prompt)
	if err != nil {
		Die(fmt.Sprintf("failed to read stdin for confirmation: %v", err))
	}
	return yesNo
}

// Die will print a message and exit (non-zero)
func Die(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

// IsPipe will indicate if we're receiving pipe input
func (a *DefaultCommand) IsPipe() bool {
	return platform.IsInputFromPipe()
}

// Input will read user input
func (a *DefaultCommand) Input(interactive, isPassword bool, prompt string) ([]byte, error) {
	return platform.GetUserInput(interactive, isPassword, prompt)
}

// Read will read a configuration file
func (c ConfigLoader) Read(file string) (io.Reader, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// Check will check if a configuration file is valid for reading
func (c ConfigLoader) Check(file string) bool {
	return platform.PathExists(file)
}

// Home will read the user's home directory
func (c ConfigLoader) Home() (string, error) {
	return os.UserHomeDir()
}
