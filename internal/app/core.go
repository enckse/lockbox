// Package app common objects
package app

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"git.sr.ht/~enckse/lockbox/internal/backend"
	"git.sr.ht/~enckse/lockbox/internal/platform"
)

type (
	// CommandOptions define how commands operate as an application
	CommandOptions interface {
		Confirm(string) bool
		Args() []string
		Transaction() *backend.Transaction
		Writer() io.Writer
	}

	// UserInputOptions handle user inputs (e.g. password entry)
	UserInputOptions interface {
		CommandOptions
		IsPipe() bool
		Input(bool) ([]byte, error)
	}

	// DefaultCommand is the default CLI app type for actual execution
	DefaultCommand struct {
		tx   *backend.Transaction
		args []string
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

// SetArgs allow updating the command args
func (a *DefaultCommand) SetArgs(args ...string) {
	a.args = args
}

// IsPipe will indicate if we're receiving pipe input
func (a *DefaultCommand) IsPipe() bool {
	return platform.IsInputFromPipe()
}

// ReadLine handles a single stdin read
func (a DefaultCommand) ReadLine() (string, error) {
	return platform.Stdin(true)
}

// Password is how a keyer gets the user's password for rekey
func (a DefaultCommand) Password() (string, error) {
	return platform.ReadInteractivePassword()
}

// Input will read user input
func (a *DefaultCommand) Input(interactive bool) ([]byte, error) {
	return platform.GetUserInputPassword(interactive)
}

func generatePrinter[T any](w io.Writer, isFilter bool, filter string, toMatch, toDisplay func(T) string) (func(T), error) {
	printer := func(p T) {
		fmt.Fprintf(w, "%s\n", toDisplay(p))
	}
	finder := printer
	if isFilter {
		re, err := regexp.Compile(filter)
		if err != nil {
			return nil, err
		}
		finder = func(p T) {
			if re.MatchString(toMatch(p)) {
				printer(p)
			}
		}
	}
	return finder, nil
}
