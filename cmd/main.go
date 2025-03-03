// provides the binary runs or calls lockbox app.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"codeberg.org/enckse/lockbox/internal/app"
	"codeberg.org/enckse/lockbox/internal/app/commands"
	"codeberg.org/enckse/lockbox/internal/config"
	"codeberg.org/enckse/lockbox/internal/platform"
	"codeberg.org/enckse/lockbox/internal/platform/clip"
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
		fmt.Printf("version: %s\n", version)
		return true, nil
	case commands.Clear:
		return true, clearClipboard()
	}
	return false, nil
}

func run() error {
	for _, p := range config.NewConfigFiles() {
		if platform.PathExists(p) {
			if err := config.LoadConfigFile(p); err != nil {
				return err
			}
			break
		}
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
	switch command {
	case commands.ReKey:
		return app.ReKey(p)
	case commands.List:
		return app.List(p)
	case commands.Move:
		return app.Move(p)
	case commands.Insert, commands.MultiLine:
		mode := app.SingleLineInsert
		if command == commands.MultiLine {
			mode = app.MultiLineInsert
		}
		return app.Insert(p, mode)
	case commands.Remove:
		return app.Remove(p)
	case commands.JSON:
		return app.JSON(p)
	case commands.Show, commands.Clip:
		return app.ShowClip(p, command == commands.Show)
	case commands.Conv:
		return app.Conv(p)
	case commands.TOTP:
		args, err := app.NewTOTPArguments(sub, config.EnvTOTPEntry.Get())
		if err != nil {
			return err
		}
		if args.Mode == app.InsertTOTPMode {
			p.SetArgs(args.Entry)
			return app.Insert(p, app.TOTPInsert)
		}
		return args.Do(app.NewDefaultTOTPOptions(p))
	case commands.PasswordGenerate:
		return app.GeneratePassword(p)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func clearClipboard() error {
	var idx int64
	val, err := platform.Stdin(false)
	if err != nil {
		return err
	}
	clipboard, err := clip.New()
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
