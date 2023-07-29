// Package app handles informational requests
package app

import (
	"errors"
	"fmt"
	"io"
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
		results, err := Usage(isAdvanced)
		if err != nil {
			return nil, err
		}
		return results, nil
	case EnvCommand, BashCommand, ZshCommand:
		defaultFlag := BashDefaultsCommand
		secondFlag := ""
		isEnv := command == EnvCommand
		if isEnv {
			defaultFlag = EnvDefaultsCommand
			secondFlag = EnvShortCommand
		}
		defaults, short, err := getInfoDefault(args, defaultFlag, secondFlag)
		if err != nil {
			return nil, err
		}
		if isEnv {
			if short {
				return config.Environ(), nil
			}
			return config.ListEnvironmentVariables(!defaults), nil
		}
		return GenerateCompletions(command == BashCommand, defaults)
	}
	return nil, nil
}

func getInfoDefault(args []string, possibleArg, otherArg string) (bool, bool, error) {
	first := false
	second := false
	invalid := false
	switch len(args) {
	case 0:
		break
	case 1:
		arg := args[0]
		if arg == possibleArg {
			first = true
		} else {
			invalid = true
			if otherArg != "" && arg == otherArg {
				invalid = false
				second = true
			}
		}
	default:
		invalid = true
	}
	if invalid {
		return false, false, errors.New("invalid argument")
	}
	return first, second, nil
}
