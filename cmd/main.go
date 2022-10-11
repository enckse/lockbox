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
	totpCommand    = "totp"
	hashCommand    = "hash"
	clearCommand   = "clear"
	clipCommand    = "clip"
	findCommand    = "find"
	insertCommand  = "insert"
	listCommand    = "ls"
	moveCommand    = "mv"
	showCommand    = "show"
	versionCommand = "version"
	helpCommand    = "help"
	removeCommand  = "rm"
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

func printSubCommand(name, desc string) {
	printCommandText("        ", name, desc)
}

func printCommand(name, desc string) {
	printCommandText("  ", name, desc)
}

func printCommandText(offset, name, desc string) {
	fmt.Printf("%s%10s    %s\n", offset, name, desc)
}

func printUsage() {
	fmt.Println("lb usage:")
	printCommand(clipCommand, "copy the entry's value into the clipboard")
	printCommand(findCommand, "perform a simplistic text search over the entry keys")
	printCommand(helpCommand, "show this usage information")
	printCommand(insertCommand, "insert a new entry into the store")
	printCommand(listCommand, "list entries")
	printCommand(moveCommand, "move an entry from one location to another with the store")
	printCommand(removeCommand, "remove an entry from the store")
	printCommand(showCommand, "show the entry's value")
	printCommand(totpCommand, "display an updating totp generated code")
	printSubCommand(totp.ClipCommand, "copy totp code to clipboard")
	printSubCommand(totp.ListCommand, "list entries with totp settings")
	printSubCommand(totp.OnceCommand, "display the first generated code")
	printSubCommand(totp.ShortCommand, "display the first generated code with no details")
	printCommand(versionCommand, "display version information")
	os.Exit(0)
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
	case listCommand, findCommand:
		opts := backend.QueryOptions{}
		opts.Mode = backend.ListMode
		if command == findCommand {
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
	case versionCommand:
		fmt.Printf("version: %s\n", strings.TrimSpace(version))
	case moveCommand:
		if len(args) != 4 {
			return newError("mv requires src and dst", errors.New("src/dst required"))
		}
		src := args[2]
		dst := args[3]
		srcExists, err := t.Get(src, backend.SecretValue)
		if err != nil {
			return newError("unable to get source object", errors.New("failed to get source"))
		}
		if srcExists == nil {
			return newError("no source object found", errors.New("source object required"))
		}
		dstExists, err := t.Get(dst, backend.BlankValue)
		if err != nil {
			return newError("unable to get destination object", errors.New("failed to get destination"))
		}
		if dstExists != nil {
			if !confirm("overwrite destination") {
				return nil
			}
		}
		if err := t.Move(*srcExists, dst); err != nil {
			return newError("unable to move object", err)
		}
	case insertCommand:
		multi := false
		idx := 2
		switch len(args) {
		case 2:
			return newError("insert missing required arguments", errors.New("entry required"))
		case 3:
		case 4:
			if args[2] != "-multi" {
				return newError("multi-line insert must be after 'insert'", errors.New("invalid command"))
			}
			multi = true
			idx = 3
		default:
			return newError("too many arguments", errors.New("insert can only perform one operation"))
		}
		isPipe := inputs.IsInputFromPipe()
		entry := args[idx]
		existing, err := t.Get(entry, backend.BlankValue)
		if err != nil {
			return newError("unable to insert entry", err)
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
			return newError("invalid input", err)
		}
		p := strings.TrimSpace(string(password))
		if err := t.Insert(entry, p); err != nil {
			return newError("failed to insert", err)
		}
		fmt.Println("")
	case removeCommand:
		if len(args) != 3 {
			return newError("rm requires a single entry", errors.New("missing argument"))
		}
		deleting := args[2]
		postfixRemove := "y"
		existings, err := t.MatchPath(deleting)
		if err != nil {
			return newError("unable to get entity to delete", err)
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
				return newError("unable to remove entry", err)
			}
		}
	case showCommand, clipCommand:
		if len(args) != 3 {
			return newError("requires a single entry", fmt.Errorf("%s missing argument", command))
		}
		entry := args[2]
		clipboard := platform.Clipboard{}
		isShow := command == showCommand
		if !isShow {
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
	case helpCommand:
		printUsage()
	default:
		if len(args) < 2 {
			return newError("command missing required arguments", fmt.Errorf("%s missing argument", command))
		}
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

func hashText(args []string) error {
	if len(args) == 0 {
		return errors.New("git diff requires a file")
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
		exit("failed to get response", err)
	}
	return yesNo
}
