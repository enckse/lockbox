// Package app handles informational requests
package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	case HelpCommand:
		if len(args) > 1 {
			return nil, errors.New("invalid help command")
		}
		isAdvanced := false
		if len(args) == 1 {
			switch args[0] {
			case HelpAdvancedCommand:
				isAdvanced = true
			default:
				return nil, errors.New("invalid help option")
			}
		}
		exe, err := exeName()
		if err != nil {
			return nil, err
		}
		results, err := Usage(isAdvanced, exe)
		if err != nil {
			return nil, err
		}
		return results, nil
	case EnvCommand:
		if len(args) != 0 {
			return nil, errors.New("invalid env command")
		}
		return config.Environ(), nil
	case CompletionsCommand:
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
		switch shell {
		case CompletionsZshCommand, CompletionsBashCommand, CompletionsFishCommand:
			break
		default:
			return nil, fmt.Errorf("unknown completion type: %s", shell)
		}
		return GenerateCompletions(shell, exe)
	}
	return nil, nil
}
