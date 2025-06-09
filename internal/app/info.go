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

	"git.sr.ht/~enckse/lockbox/internal/app/commands"
	"git.sr.ht/~enckse/lockbox/internal/app/completions"
	"git.sr.ht/~enckse/lockbox/internal/app/help"
	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
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
		results, err := help.Usage(isAdvanced, commands.Executable)
		if err != nil {
			return nil, err
		}
		return results, nil
	case commands.Env:
		var set []string
		switch len(args) {
		case 0:
		case 1:
			set = []string{args[0]}
		default:
			return nil, errors.New("invalid env command, too many arguments")
		}
		var results []string
		for _, item := range store.List(set...) {
			value := fmt.Sprintf("%s=%v", item.Key, item.Value)
			results = append(results, value)
		}
		if len(results) == 0 {
			results = []string{""}
		}
		return results, nil
	case commands.Completions:
		shell := ""
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
		return completions.Generate(shell, commands.Executable)
	}
	return nil, nil
}
