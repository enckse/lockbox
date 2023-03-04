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
	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
	"github.com/enckse/lockbox/internal/totp"
	o "github.com/enckse/pgl/os"
)

//go:embed "vers.txt"
var version string

func main() {
	if err := run(); err != nil {
		o.Die(err)
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
	p, err := app.NewDefaultCommand(sub)
	if err != nil {
		return err
	}
	switch command {
	case cli.ReKeyCommand:
		if p.Confirm("proceed with rekey") {
			return p.Transaction().ReKey()
		}
	case cli.ListCommand, cli.FindCommand:
		return app.ListFind(p, command == cli.FindCommand)
	case cli.MoveCommand:
		return app.Move(p)
	case cli.InsertCommand:
		insertArgs, err := app.ReadArgs(p)
		if err != nil {
			return err
		}
		return insertArgs.Do(p)
	case cli.RemoveCommand:
		return app.Remove(p)
	case cli.StatsCommand:
		return app.Stats(p)
	case cli.ShowCommand, cli.ClipCommand:
		return app.ShowClip(p, command == cli.ShowCommand)
	case cli.HashCommand:
		return app.Hash(p)
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
