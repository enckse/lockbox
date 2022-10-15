// provides the binary runs or calls lockbox commands.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/totp"
)

const (
	totpCommand        = "totp"
	hashCommand        = "hash"
	clearCommand       = "clear"
	clipCommand        = "clip"
	findCommand        = "find"
	insertCommand      = "insert"
	listCommand        = "ls"
	moveCommand        = "mv"
	showCommand        = "show"
	versionCommand     = "version"
	helpCommand        = "help"
	removeCommand      = "rm"
	envCommand         = "env"
	insertMultiCommand = "-multi"
)

var (
	//go:embed "vers.txt"
	version string
)

type (
	callbackFunction func([]string) error
	wrappedError     struct {
		message string
	}
)

func printSubCommand(parent, name, args, desc string) {
	printCommandText(args, fmt.Sprintf("%s %s", parent, name), desc)
}

func printCommand(name, args, desc string) {
	printCommandText(args, name, desc)
}

func printCommandText(args, name, desc string) {
	arguments := ""
	if len(args) > 0 {
		arguments = fmt.Sprintf("[%s]", args)
	}
	fmt.Printf("  %-15s %-10s    %s\n", name, arguments, desc)
}

func printUsage() {
	fmt.Println("lb usage:")
	printCommand(clipCommand, "entry", "copy the entry's value into the clipboard")
	printCommand(envCommand, "", "display environment variable information")
	printCommand(findCommand, "criteria", "perform a simplistic text search over the entry keys")
	printCommand(helpCommand, "", "show this usage information")
	printCommand(insertCommand, "entry", "insert a new entry into the store")
	printSubCommand(insertCommand, insertMultiCommand, "entry", "insert a multi-line entry")
	printCommand(listCommand, "", "list entries")
	printCommand(moveCommand, "src dst", "move an entry from one location to another with the store")
	printCommand(removeCommand, "entry", "remove an entry from the store")
	printCommand(showCommand, "entry", "show the entry's value")
	printCommand(totpCommand, "entry", "display an updating totp generated code")
	printSubCommand(totpCommand, totp.ClipCommand, "entry", "copy totp code to clipboard")
	printSubCommand(totpCommand, totp.ListCommand, "", "list entries with totp settings")
	printSubCommand(totpCommand, totp.OnceCommand, "entry", "display the first generated code")
	printSubCommand(totpCommand, totp.ShortCommand, "entry", "display the first generated code with no details")
	printCommand(versionCommand, "", "display version information")
}

func internalCallback(name string) callbackFunction {
	switch name {
	case totpCommand:
		return totp.Call
	case hashCommand:
		return hashText
	case clearCommand:
		return clearClipboard
	}
	return nil
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	if err := run(); err != nil {
		exit(err)
	}
}

func processInfoCommands(command string, args []string) (bool, error) {
	switch command {
	case helpCommand:
		printUsage()
	case versionCommand:
		fmt.Printf("version: %s\n", strings.TrimSpace(version))
	case envCommand:
		printValues := true
		invalid := false
		switch len(args) {
		case 2:
			break
		case 3:
			if args[2] == "-defaults" {
				printValues = false
			} else {
				invalid = true
			}
		default:
			invalid = true
		}
		if invalid {
			return false, errors.New("invalid argument")
		}
		inputs.ListEnvironmentVariables(printValues)
	default:
		return false, nil
	}
	return true, nil
}

func wrapped(message string, err error) wrappedError {
	return wrappedError{message: fmt.Sprintf("%s (%v)", message, err)}
}

func (w wrappedError) Error() string {
	return w.message
}

func run() error {
	args := os.Args
	if len(args) < 2 {
		return errors.New("requires subcommand")
	}
	command := args[1]
	ok, err := processInfoCommands(command, args)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	t, err := backend.NewTransaction()
	if err != nil {
		return wrapped("unable to build transaction model", err)
	}
	switch command {
	case listCommand, findCommand:
		opts := backend.QueryOptions{}
		opts.Mode = backend.ListMode
		if command == findCommand {
			opts.Mode = backend.FindMode
			if len(args) < 3 {
				return errors.New("find requires search term")
			}
			opts.Criteria = args[2]
		}
		e, err := t.QueryCallback(opts)
		if err != nil {
			return wrapped("unable to list files", err)
		}
		for _, f := range e {
			fmt.Println(f.Path)
		}
	case moveCommand:
		if len(args) != 4 {
			return errors.New("src/dst required for move")
		}
		src := args[2]
		dst := args[3]
		srcExists, err := t.Get(src, backend.SecretValue)
		if err != nil {
			return errors.New("unable to get source entry")
		}
		if srcExists == nil {
			return errors.New("no source object found")
		}
		dstExists, err := t.Get(dst, backend.BlankValue)
		if err != nil {
			return errors.New("unable to get destination object")
		}
		if dstExists != nil {
			if !confirm("overwrite destination") {
				return nil
			}
		}
		if err := t.Move(*srcExists, dst); err != nil {
			return wrapped("unable to move object", err)
		}
	case insertCommand:
		multi := false
		idx := 2
		switch len(args) {
		case 2:
			return errors.New("insert requires an entry")
		case 3:
		case 4:
			if args[2] != insertMultiCommand {
				return errors.New("unknown argument")
			}
			multi = true
			idx = 3
		default:
			return errors.New("too many arguments")
		}
		isPipe := inputs.IsInputFromPipe()
		entry := args[idx]
		existing, err := t.Get(entry, backend.BlankValue)
		if err != nil {
			return wrapped("unable to check for existing entry", err)
		}
		if existing != nil {
			if !isPipe {
				if !confirm("overwrite existing") {
					return nil
				}
			}
		}
		password, err := inputs.GetUserInputPassword(isPipe, multi)
		if err != nil {
			return wrapped("invalid input", err)
		}
		p := strings.TrimSpace(string(password))
		if err := t.Insert(entry, p); err != nil {
			return wrapped("failed to insert", err)
		}
		fmt.Println("")
	case removeCommand:
		if len(args) != 3 {
			return errors.New("remove requires an entry")
		}
		deleting := args[2]
		postfixRemove := "y"
		existings, err := t.MatchPath(deleting)
		if err != nil {
			return wrapped("unable to get entry", err)
		}

		if len(existings) > 1 {
			postfixRemove = "ies"
			fmt.Println("selected entities:")
			for _, e := range existings {
				fmt.Printf(" %s\n", e.Path)
			}
			fmt.Println("")
		}
		if confirm(fmt.Sprintf("delete entr%s", postfixRemove)) {
			if err := t.RemoveAll(existings); err != nil {
				return wrapped("unable to remove entry", err)
			}
		}
	case showCommand, clipCommand:
		if len(args) != 3 {
			return errors.New("entry required")
		}
		entry := args[2]
		clipboard := platform.Clipboard{}
		isShow := command == showCommand
		if !isShow {
			clipboard, err = platform.NewClipboard()
			if err != nil {
				return wrapped("unable to get clipboard", err)
			}
		}
		existing, err := t.Get(entry, backend.SecretValue)
		if err != nil {
			return wrapped("unable to get entry", err)
		}
		if existing == nil {
			return errors.New("entry not found")
		}
		if isShow {
			fmt.Println(existing.Value)
			return nil
		}
		if err := clipboard.CopyTo(existing.Value); err != nil {
			return wrapped("clipboard operation failed", err)
		}
	default:
		if len(args) < 2 {
			return errors.New("missing required arguments")
		}
		a := args[2:]
		callback := internalCallback(command)
		if callback != nil {
			if err := callback(a); err != nil {
				return wrapped(fmt.Sprintf("%s command failure", command), err)
			}
			return nil
		}
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}

func hashText(args []string) error {
	if len(args) == 0 {
		return errors.New("hash requires a file")
	}
	t, err := backend.Load(args[len(args)-1])
	if err != nil {
		return err
	}
	e, err := t.QueryCallback(backend.QueryOptions{Mode: backend.ListMode, Values: backend.HashedValue})
	if err != nil {
		return err
	}
	for _, item := range e {
		fmt.Printf("%s:\nhash:%s\n", item.Path, item.Value)
	}
	return nil
}

func clearClipboard(args []string) error {
	idx := 0
	val, err := inputs.Stdin(false)
	if err != nil {
		return err
	}
	clipboard, err := platform.NewClipboard()
	if err != nil {
		return err
	}
	pCmd, pArgs := clipboard.Args(false)
	val = strings.TrimSpace(val)
	for idx < clipboard.MaxTime {
		idx++
		time.Sleep(1 * time.Second)
		out, err := exec.Command(pCmd, pArgs...).Output()
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(out)) != val {
			return nil
		}
	}
	return clipboard.CopyTo("")
}

func confirm(prompt string) bool {
	yesNo, err := inputs.ConfirmYesNoPrompt(prompt)
	if err != nil {
		exit(wrapped("failed to get response", err))
	}
	return yesNo
}
