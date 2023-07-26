// provides the binary runs or calls lockbox app.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/platform"
)

var version string

func main() {
	if err := run(); err != nil {
		app.Die(err.Error())
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
	case app.VersionCommand:
		fmt.Printf("version: %s\n", version)
		return true, nil
	case app.ClearCommand:
		return true, clearClipboard()
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
	case app.ReKeyCommand:
		keyer, err := app.NewDefaultKeyer()
		if err != nil {
			return err
		}
		return app.ReKey(p, keyer)
	case app.ListCommand:
		return app.List(p)
	case app.MoveCommand:
		return app.Move(p)
	case app.InsertCommand, app.MultiLineCommand:
		mode := app.SingleLineInsert
		if command == app.MultiLineCommand {
			mode = app.MultiLineInsert
		}
		return app.Insert(p, mode)
	case app.RemoveCommand:
		return app.Remove(p)
	case app.JSONCommand:
		return app.JSON(p)
	case app.ShowCommand, app.ClipCommand:
		return app.ShowClip(p, command == app.ShowCommand)
	case app.ConvCommand:
		return app.Conv(p)
	case app.TOTPCommand:
		args, err := app.NewTOTPArguments(sub, inputs.TOTPToken())
		if err != nil {
			return err
		}
		if args.Mode == app.InsertTOTPMode {
			p.SetArgs(args.Entry)
			return app.Insert(p, app.TOTPInsert)
		}
		return args.Do(app.NewDefaultTOTPOptions(p))
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func clearClipboard() error {
	idx := 0
	val, err := platform.Stdin(false)
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
