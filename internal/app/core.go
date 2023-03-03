// Package app runs the commands/ui
package app

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/commands"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/totp"
	"github.com/enckse/lockbox/internal/util"
)

//go:embed "vers.txt"
var version string

func handleEarly(command string, args []string) (bool, error) {
	ok, err := commands.Info(os.Stdout, command, args)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	switch command {
	case cli.VersionCommand:
		fmt.Printf("version: %s\n", version)
		return true, nil
	case cli.TOTPCommand:
		return true, totp.Call(args)
	case cli.HashCommand:
		return true, commands.Hash(os.Stdout, args)
	case cli.ClearCommand:
		return true, clearClipboard(args)
	}
	return false, nil
}

func wrapped(message string, err error) error {
	return fmt.Errorf("%s (%w)", message, err)
}

// Run invokes the app
func Run() error {
	args := os.Args
	if len(args) < 2 {
		return errors.New("requires subcommand")
	}
	command := args[1]
	sub := args[2:]
	ok, err := handleEarly(command, sub)
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
	case cli.ReKeyCommand:
		if confirm("proceed with rekey") {
			return t.ReKey()
		}
	case cli.ListCommand, cli.FindCommand:
		return commands.ListFind(t, os.Stdout, command == cli.FindCommand, sub)
	case cli.MoveCommand:
		return commands.Move(t, sub, confirm)
	case cli.InsertCommand:
		multi := false
		isTOTP := false
		idx := 2
		switch len(args) {
		case 2:
			return errors.New("insert requires an entry")
		case 3:
		case 4:
			opt := args[2]
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
			idx = 3
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
		if !isPipe {
			fmt.Println()
		}
	case cli.RemoveCommand:
		return commands.Remove(os.Stdout, t, sub, confirm)
	case cli.StatsCommand:
		return commands.Stats(os.Stdout, t, sub)
	case cli.ShowCommand, cli.ClipCommand:
		if len(args) != 3 {
			return errors.New("entry required")
		}
		entry := args[2]
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
		return fmt.Errorf("unknown command: %s", command)
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
	pCmd, pArgs, valid := clipboard.Args(false)
	if !valid {
		return nil
	}
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
		util.Dief("failed to read stdin for confirmation: %v", err)
	}
	return yesNo
}
