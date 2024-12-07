// Package app handles informational requests
package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/seanenck/lockbox/internal/app/commands"
	"github.com/seanenck/lockbox/internal/app/completions"
	"github.com/seanenck/lockbox/internal/app/help"
	"github.com/seanenck/lockbox/internal/config"
)

// Info will report help/bash/env details
func Info(w io.Writer, command string, args []string) (bool, error) {
	i, err := info(command, args)
	if err != nil {
		return false, err
	}
	if len(i) > 0 {
		fmt.Fprintf(w, "%s\n", strings.Join(i, "\n"))
		return true, nil
	}
	return false, nil
}

func exeName() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Base(exe), nil
}

func info(command string, args []string) ([]string, error) {
	switch command {
	case commands.Help:
		if len(args) > 1 {
			return nil, errors.New("invalid help command")
		}
		isAdvanced := false
		if len(args) == 1 {
			switch args[0] {
			case commands.HelpAdvanced:
				isAdvanced = true
			case commands.HelpConfig:
				data, err := config.DefaultTOML()
				if err != nil {
					return nil, err
				}
				return []string{data}, nil
			default:
				return nil, errors.New("invalid help option")
			}
		}
		exe, err := exeName()
		if err != nil {
			return nil, err
		}
		results, err := help.Usage(isAdvanced, exe)
		if err != nil {
			return nil, err
		}
		return results, nil
	case commands.Env:
		var set []string
		switch len(args) {
		case 0:
		case 1:
			sub := args[0]
			if sub == commands.Completions {
				set = completions.NewConditionals().Exported
			} else {
				set = []string{sub}
			}
		default:
			return nil, errors.New("invalid env command, too many arguments")
		}
		env := config.Environ(set...)
		if len(env) == 0 {
			env = []string{""}
		}
		return env, nil
	case commands.Completions:
		shell := ""
		exe, err := exeName()
		if err != nil {
			return nil, err
		}
		switch len(args) {
		case 0:
			shell = filepath.Base(os.Getenv("SHELL"))
		case 1:
			shell = args[0]
		default:
			return nil, errors.New("invalid completions subcommand")
		}
		if !slices.Contains(commands.CompletionTypes, shell) {
			return nil, fmt.Errorf("unknown completion type: %s", shell)
		}
		return completions.Generate(shell, exe)
	}
	return nil, nil
}
