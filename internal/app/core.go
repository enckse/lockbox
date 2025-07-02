// Package app common objects
package app

import (
	"fmt"
	"io"
	"os"

	"git.sr.ht/~enckse/lockbox/internal/kdbx"
	"git.sr.ht/~enckse/lockbox/internal/platform"
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
		Input(bool, string) ([]byte, error)
	}

	// DefaultCommand is the default CLI app type for actual execution
	DefaultCommand struct {
		tx   *kdbx.Transaction
		args []string
	}
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
func (a *DefaultCommand) Input(interactive bool, prompt string) ([]byte, error) {
	return platform.GetUserInput(interactive, prompt)
}
