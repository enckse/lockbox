// Package app common objects
package app

import (
	"io"
	"os"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/pgl/os/exit"
)

type (
	// CommandOptions define how commands operate as an application
	CommandOptions interface {
		Confirm(string) bool
		Args() []string
		Transaction() *backend.Transaction
		Writer() io.Writer
	}

	// DefaultCommand is the default CLI app type for actual execution
	DefaultCommand struct {
		args []string
		tx   *backend.Transaction
	}
)

// NewDefaultCommand creates a new app command
func NewDefaultCommand(args []string) (*DefaultCommand, error) {
	t, err := backend.NewTransaction()
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
func (a *DefaultCommand) Transaction() *backend.Transaction {
	return a.tx
}

// Confirm will confirm with the user (dying if something abnormal happens)
func (a *DefaultCommand) Confirm(prompt string) bool {
	yesNo, err := inputs.ConfirmYesNoPrompt(prompt)
	if err != nil {
		exit.Dief("failed to read stdin for confirmation: %v", err)
	}
	return yesNo
}

// SetArgs allow updating the command args
func (a *DefaultCommand) SetArgs(args ...string) {
	a.args = args
}

// IsPipe will indicate if we're receiving pipe input
func (a *DefaultCommand) IsPipe() bool {
	return inputs.IsInputFromPipe()
}

// Input will read user input
func (a *DefaultCommand) Input(pipe, multi bool) ([]byte, error) {
	return inputs.GetUserInputPassword(pipe, multi)
}
