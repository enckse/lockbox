// Package program is the underlying full CLI representation to run lb
package program

import (
	_ "embed"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/totp"
)

var (
	//go:embed "vers.txt"
	version string
)

type (
	// ConfirmFunction is used to confirm the UI selections
	ConfirmFunction func(string) (bool, error)
	// ExitFunction is used to exit (unexpectedly)
	ExitFunction     func(error)
	callbackFunction func([]string) error
	// Program wraps calling/running the underlying core functionality
	Program struct {
		confirm func(string) bool
	}
)

// NewProgram creates a new runnable program
func NewProgram(confirmation ConfirmFunction, exit ExitFunction) (*Program, error) {
	if confirmation == nil {
		return nil, errors.New("confirmation function must be set")
	}
	if exit == nil {
		return nil, errors.New("exit function must be set")
	}
	return &Program{confirm: func(arg string) bool {
		c, err := confirmation(arg)
		if err != nil {
			exit(err)
		}
		return c
	}}, nil
}

func internalCallback(name string) callbackFunction {
	switch name {
	case cli.TOTPCommand:
		return totp.Call
	case cli.HashCommand:
		return hashText
	case cli.ClearCommand:
		return clearClipboard
	}
	return nil
}

func getInfoDefault(args []string, possibleArg string) (bool, error) {
	defaults := false
	invalid := false
	switch len(args) {
	case 1:
		break
	case 2:
		if args[1] == possibleArg {
			defaults = true
		} else {
			invalid = true
		}
	default:
		invalid = true
	}
	if invalid {
		return false, errors.New("invalid argument")
	}
	return defaults, nil
}

func processInfoCommands(command string, args []string) ([]string, error) {
	switch command {
	case cli.HelpCommand:
		return cli.Usage()
	case cli.VersionCommand:
		return []string{fmt.Sprintf("version: %s", strings.TrimSpace(version))}, nil
	case cli.EnvCommand, cli.BashCommand:
		defaultFlag := cli.BashDefaultsCommand
		isEnv := command == cli.EnvCommand
		if isEnv {
			defaultFlag = cli.EnvDefaultsCommand
		}
		defaults, err := getInfoDefault(args, defaultFlag)
		if err != nil {
			return nil, err
		}
		if isEnv {
			return inputs.ListEnvironmentVariables(!defaults), nil
		}
		return cli.BashCompletions(defaults)
	}
	return nil, nil
}

func wrapped(message string, err error) error {
	return fmt.Errorf("%s (%v)", message, err)
}

// Run will run the program
func (p *Program) Run(inArgs []string) error {
	if p.confirm == nil {
		return errors.New("invalid program, not setup properly")
	}
	args := inArgs
	if len(args) < 1 {
		return errors.New("requires subcommand")
	}
	command := args[0]
	info, err := processInfoCommands(command, args)
	if err != nil {
		return err
	}
	if info != nil {
		fmt.Println(strings.Join(info, "\n"))
		return nil
	}
	t, err := backend.NewTransaction()
	if err != nil {
		return wrapped("unable to build transaction model", err)
	}
	switch command {
	case cli.ListCommand, cli.FindCommand:
		opts := backend.QueryOptions{}
		opts.Mode = backend.ListMode
		if command == cli.FindCommand {
			opts.Mode = backend.FindMode
			if len(args) < 2 {
				return errors.New("find requires search term")
			}
			opts.Criteria = args[1]
		}
		e, err := t.QueryCallback(opts)
		if err != nil {
			return wrapped("unable to list files", err)
		}
		for _, f := range e {
			fmt.Println(f.Path)
		}
	case cli.MoveCommand:
		if len(args) != 3 {
			return errors.New("src/dst required for move")
		}
		src := args[1]
		dst := args[2]
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
			if !p.confirm("overwrite destination") {
				return nil
			}
		}
		if err := t.Move(*srcExists, dst); err != nil {
			return wrapped("unable to move object", err)
		}
	case cli.InsertCommand:
		multi := false
		isTOTP := false
		idx := 1
		switch len(args) {
		case 1:
			return errors.New("insert requires an entry")
		case 2:
		case 3:
			opt := args[1]
			switch opt {
			case cli.InsertMultiCommand:
				multi = true
			case cli.InsertTOTPCommand:
				off, err := inputs.IsNoTOTP()
				if err != nil {
					return err
				}
				if off {
					return totp.ErrNoTOTP
				}
				isTOTP = true
			default:
				return errors.New("unknown argument")
			}
			multi = true
			idx = 2
		default:
			return errors.New("too many arguments")
		}
		isPipe := inputs.IsInputFromPipe()
		entry := args[idx]
		if isTOTP {
			totpToken := inputs.TOTPToken()
			if !strings.HasSuffix(entry, backend.NewSuffix(totpToken)) {
				entry = backend.NewPath(entry, totpToken)
			}
		}
		existing, err := t.Get(entry, backend.BlankValue)
		if err != nil {
			return wrapped("unable to check for existing entry", err)
		}
		if existing != nil {
			if !isPipe {
				if !p.confirm("overwrite existing") {
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
	case cli.RemoveCommand:
		if len(args) != 2 {
			return errors.New("remove requires an entry")
		}
		deleting := args[1]
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
		if p.confirm(fmt.Sprintf("delete entr%s", postfixRemove)) {
			if err := t.RemoveAll(existings); err != nil {
				return wrapped("unable to remove entry", err)
			}
		}
	case cli.ShowCommand, cli.ClipCommand:
		if len(args) != 2 {
			return errors.New("entry required")
		}
		entry := args[1]
		clipboard := platform.Clipboard{}
		isShow := command == cli.ShowCommand
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
		if len(args) < 1 {
			return errors.New("missing required arguments")
		}
		a := args[1:]
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
