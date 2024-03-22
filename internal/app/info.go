// Package app handles informational requests
package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/enckse/lockbox/internal/config"
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
			if args[0] == HelpAdvancedCommand {
				isAdvanced = true
			} else {
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
	case BashCommand, ZshCommand, FishCommand:
		if len(args) > 1 {
			return nil, fmt.Errorf("invalid %s command", command)
		}
		isHelp := false
		if len(args) == 1 {
			if args[0] == CompletionHelpCommand {
				isHelp = true
			} else {
				return nil, fmt.Errorf("invalid %s subcommand", command)
			}
		}
		exe, err := exeName()
		if err != nil {
			return nil, err
		}
		return GenerateCompletions(command, isHelp, exe)
	}
	return nil, nil
}
