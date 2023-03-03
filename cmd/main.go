// provides the binary runs or calls lockbox app.
package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/totp"
	"github.com/enckse/lockbox/internal/util"
)

//go:embed "vers.txt"
var version string

func main() {
	if err := run(); err != nil {
		util.Die(err)
	}
}

func handleEarly(command string, args []string) (bool, error) {
	ok, err := app.Info(os.Stdout, command, args)
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
		return true, app.Hash(os.Stdout, args)
	case cli.ClearCommand:
		return true, clearClipboard(args)
	}
	return false, nil
}

func run() error {
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
		return fmt.Errorf("unable to build transaction model: %w", err)
	}
	switch command {
	case cli.ReKeyCommand:
		if confirm("proceed with rekey") {
			return t.ReKey()
		}
	case cli.ListCommand, cli.FindCommand:
		return app.ListFind(t, os.Stdout, command == cli.FindCommand, sub)
	case cli.MoveCommand:
		return app.Move(t, sub, confirm)
	case cli.InsertCommand:
		return app.Insert(os.Stdout, t, sub, confirm)
	case cli.RemoveCommand:
		return app.Remove(os.Stdout, t, sub, confirm)
	case cli.StatsCommand:
		return app.Stats(os.Stdout, t, sub)
	case cli.ShowCommand, cli.ClipCommand:
		return app.ShowClip(os.Stdout, t, command == cli.ShowCommand, sub)
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
