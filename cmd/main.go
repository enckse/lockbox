// provides the binary runs or calls lockbox commands.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/subcommands"
)

var (
	//go:embed "vers.txt"
	version string
)

type (
	callbackFunction func([]string) error
	programError     struct {
		message string
		err     error
	}
)

func getEntry(args []string, idx int) string {
	if len(args) != idx+1 {
		exit("invalid entry given", errors.New("specific entry required"))
	}
	return args[idx]
}

func internalCallback(name string) callbackFunction {
	switch name {
	case "hash":
		return subcommands.Hashed
	case "totp":
		return subcommands.TOTP
	}
	return nil
}

func exit(message string, err error) {
	msg := message
	if err != nil {
		msg = fmt.Sprintf("%s (%v)", msg, err)
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func newError(message string, err error) *programError {
	return &programError{message: message, err: err}
}

func main() {
	if err := run(); err != nil {
		exit(err.message, err.err)
	}
}

func run() *programError {
	args := os.Args
	if len(args) < 2 {
		return newError("missing arguments", errors.New("requires subcommand"))
	}
	t, err := backend.NewTransaction()
	if err != nil {
		return newError("unable to build transaction model", err)
	}
	command := args[1]
	switch command {
	case "ls", "find":
		opts := backend.QueryOptions{}
		opts.Mode = backend.ListMode
		if command == "find" {
			opts.Mode = backend.FindMode
			if len(args) < 3 {
				return newError("find requires an argument to search for", errors.New("search term required"))
			}
			opts.Criteria = args[2]
		}
		e, err := t.QueryCallback(opts)
		if err != nil {
			return newError("unable to list files", err)
		}
		for _, f := range e {
			fmt.Println(f.Path)
		}
	case "version":
		fmt.Printf("version: %s\n", version)
	case "insert":
		options := cli.Arguments{}
		idx := 2
		switch len(args) {
		case 2:
			return newError("insert missing required arguments", errors.New("entry required"))
		case 3:
		case 4:
			options = cli.ParseArgs(args[2])
			if !options.Multi {
				return newError("multi-line insert must be after 'insert'", errors.New("invalid command"))
			}
			idx = 3
		default:
			return newError("too many arguments", errors.New("insert can only perform one operation"))
		}
		isPipe := inputs.IsInputFromPipe()
		entry := getEntry(args, idx)
		existing, err := t.Get(entry, backend.BlankValue)
		if err != nil {
			return newError("unable to find an exact, existing match", err)
		}
		if existing != nil {
			if !isPipe {
				if !confirm("overwrite existing") {
					return nil
				}
			}
		}
		password, err := inputs.GetUserInputPassword(isPipe, options.Multi)
		if err != nil {
			return newError("invalid input", err)
		}
		p := strings.TrimSpace(string(password))
		if err := t.Insert(entry, p, existing, len(strings.Split(p, "\n")) > 1); err != nil {
			return newError("failed to insert", err)
		}
		fmt.Println("")
	case "rm":
		deleting := getEntry(args, 2)
		existing, err := t.Get(deleting, backend.BlankValue)
		if err != nil {
			return newError("unable to get entity to delete", err)
		}
		if confirm("delete entry") {
			if err := t.Remove(existing); err != nil {
				return newError("unable to remove entry", err)
			}

		}
	case "show", "clip":
		entry := getEntry(args, 2)
		clipboard := platform.Clipboard{}
		isShow := command == "show"
		if isShow {
			clipboard, err = platform.NewClipboard()
			if err != nil {
				return newError("unable to get clipboard", err)
			}
		}
		existing, err := t.Get(entry, backend.SecretValue)
		if err != nil {
			return newError("unable to get entity", err)
		}
		if existing == nil {
			return newError("entity not found", errors.New("can not find entry"))
		}
		if isShow {
			fmt.Println(existing.Value)
			return nil
		}
		if err := clipboard.CopyTo(existing.Value); err != nil {
			return newError("clipboard failed", err)
		}

	case "clear":
		if err := subcommands.ClearClipboardCallback(); err != nil {
			return newError("failed to handle clipboard clear", err)
		}
	default:
		a := args[2:]
		callback := internalCallback(command)
		if callback != nil {
			if err := callback(a); err != nil {
				return newError(fmt.Sprintf("%s command failure", command), err)
			}
			return nil
		}
		return newError("unknown command", errors.New(command))
	}
	return nil
}

func confirm(prompt string) bool {
	yesNo, err := inputs.ConfirmYesNoPrompt(prompt)
	if err != nil {
		exit("failed to get response", err)
	}
	return yesNo
}
