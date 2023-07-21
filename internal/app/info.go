// Package app handles informational requests
package app

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/enckse/lockbox/internal/cli"
	"github.com/enckse/lockbox/internal/inputs"
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
	case cli.HelpCommand:
		if len(args) > 1 {
			return nil, errors.New("invalid help command")
		}
		isAdvanced := false
		if len(args) == 1 {
			if args[0] == cli.HelpAdvancedCommand {
				isAdvanced = true
			} else {
				return nil, errors.New("invalid help option")
			}
		}
		results, err := cli.Usage(isAdvanced)
		if err != nil {
			return nil, err
		}
		return results, nil
	case cli.EnvCommand, cli.BashCommand, cli.ZshCommand:
		defaultFlag := cli.BashDefaultsCommand
		isEnv := command == cli.EnvCommand
		if isEnv {
			defaultFlag = cli.EnvDefaultsCommand
		}
		defaults, err := getInfoDefault(args, defaultFlag)
		if err != nil {
			return nil, err
		}
		if isEnv {
			return inputs.ListEnvironmentVariables(!defaults), nil
		}
		return cli.GenerateCompletions(command == cli.BashCommand, defaults)
	}
	return nil, nil
}

func getInfoDefault(args []string, possibleArg string) (bool, error) {
	defaults := false
	invalid := false
	switch len(args) {
	case 0:
		break
	case 1:
		if args[0] == possibleArg {
			defaults = true
		} else {
			invalid = true
		}
	default:
		invalid = true
	}
	if invalid {
		return false, errors.New("invalid argument")
	}
	return defaults, nil
}
