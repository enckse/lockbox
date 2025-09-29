// provides the binary runs or calls lockbox app.
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/enckse/lockbox/internal/app"
	"github.com/enckse/lockbox/internal/app/commands"
	"github.com/enckse/lockbox/internal/config"
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
	case commands.Version:
		vers := version
		if vers == "" {
			info, ok := debug.ReadBuildInfo()
			if ok {
				vers = info.Main.Version
			}
		}
		if vers == "" {
			vers = "(development)"
		}
		fmt.Printf("version: %s\n", vers)
		return true, nil
	}
	return false, nil
}

func run() error {
	if err := config.Parse(app.ConfigLoader{}); err != nil {
		return err
	}
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
	res := commands.AllowedInReadOnly(command)
	if len(res) != 1 {
		return fmt.Errorf("%s is not allowed in read-only", command)
	}
	switch command {
	case commands.Health:
		return app.Health(p)
	case commands.ReKey:
		return app.ReKey(p)
	case commands.List, commands.Groups, commands.Fields:
		mode := app.ListEntriesMode
		switch command {
		case commands.Groups:
			mode = app.ListGroupsMode
		case commands.Fields:
			mode = app.ListFieldsMode
		}
		return app.List(p, mode)
	case commands.Unset:
		return app.Unset(p)
	case commands.Move:
		return app.Move(p)
	case commands.Insert:
		return app.Insert(p)
	case commands.Remove:
		return app.Remove(p)
	case commands.JSON:
		return app.JSON(p)
	case commands.Show, commands.Clip:
		return app.ShowClip(p, command == commands.Show)
	case commands.Conv:
		return app.Conv(p)
	case commands.TOTP:
		args, err := app.NewTOTPArguments(sub)
		if err != nil {
			return err
		}
		return args.Do(app.NewDefaultTOTPOptions(p))
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}
