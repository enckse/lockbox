// Package cli handles CLI helpers/commands
package cli

import (
	"fmt"
)

const (
	// TOTPCommand is the parent of totp and by defaults generates a rotating token
	TOTPCommand = "totp"
	// HashCommand handles hashing the data store
	HashCommand = "hash"
	// ClearCommand is a callback to manage clipboard clearing
	ClearCommand = "clear"
	// ClipCommand will copy values to the clipboard
	ClipCommand = "clip"
	// FindCommand is for simplistic searching of entries
	FindCommand = "find"
	// InsertCommand adds a value
	InsertCommand = "insert"
	// ListCommand lists all entries
	ListCommand = "ls"
	// MoveCommand will move source to destination
	MoveCommand = "mv"
	// ShowCommand will show the value in an entry
	ShowCommand = "show"
	// VersionCommand displays version information
	VersionCommand = "version"
	// HelpCommand shows usage
	HelpCommand = "help"
	// RemoveCommand removes an entry
	RemoveCommand = "rm"
	// EnvCommand shows environment information used by lockbox
	EnvCommand = "env"
	// InsertMultiCommand handles multi-line inserts
	InsertMultiCommand = "-multi"
	// TOTPClipCommand is the argument for copying totp codes to clipboard
	TOTPClipCommand = "-clip"
	// TOTPShortCommand is the argument for getting the short version of a code
	TOTPShortCommand = "-short"
	// TOTPListCommand will list the totp-enabled entries
	TOTPListCommand = "-list"
	// TOTPOnceCommand will perform like a normal totp request but not refresh
	TOTPOnceCommand = "-once"
	// EnvDefaultsCommand will display the default env variables, not those set
	EnvDefaultsCommand = "-defaults"
)

func subCommand(parent, name, args, desc string) string {
	return commandText(args, fmt.Sprintf("%s %s", parent, name), desc)
}

func command(name, args, desc string) string {
	return commandText(args, name, desc)
}

func commandText(args, name, desc string) string {
	arguments := ""
	if len(args) > 0 {
		arguments = fmt.Sprintf("[%s]", args)
	}
	return fmt.Sprintf("  %-15s %-10s    %s", name, arguments, desc)
}

// Usage return usage information
func Usage() []string {
	results := []string{"lb usage"}
	results = append(results, command(ClipCommand, "entry", "copy the entry's value into the clipboard"))
	results = append(results, command(EnvCommand, "", "display environment variable information"))
	results = append(results, command(FindCommand, "criteria", "perform a simplistic text search over the entry keys"))
	results = append(results, command(HelpCommand, "", "show this usage information"))
	results = append(results, command(InsertCommand, "entry", "insert a new entry into the store"))
	results = append(results, subCommand(InsertCommand, InsertMultiCommand, "entry", "insert a multi-line entry"))
	results = append(results, command(ListCommand, "", "list entries"))
	results = append(results, command(MoveCommand, "src dst", "move an entry from one location to another with the store"))
	results = append(results, command(RemoveCommand, "entry", "remove an entry from the store"))
	results = append(results, command(ShowCommand, "entry", "show the entry's value"))
	results = append(results, command(TOTPCommand, "entry", "display an updating totp generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPClipCommand, "entry", "copy totp code to clipboard"))
	results = append(results, subCommand(TOTPCommand, TOTPListCommand, "", "list entries with totp settings"))
	results = append(results, subCommand(TOTPCommand, TOTPOnceCommand, "entry", "display the first generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPShortCommand, "entry", "display the first generated code with no details"))
	results = append(results, command(VersionCommand, "", "display version information"))
	return results
}
