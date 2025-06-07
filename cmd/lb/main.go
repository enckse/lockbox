// provides the binary runs or calls lockbox app.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"git.sr.ht/~enckse/lockbox/internal/app"
	"git.sr.ht/~enckse/lockbox/internal/app/commands"
	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/platform"
	"git.sr.ht/~enckse/lockbox/internal/platform/clip"
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
	case commands.List, commands.Find, commands.Groups:
		return app.List(p, command == commands.Find, command == commands.Groups)
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
