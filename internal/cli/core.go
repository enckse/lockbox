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
	return fmt.Sprintf("  %-15s %-10s    %s\n", name, arguments, desc)
}

// Usage return usage information
func Usage() []string {
	fmt.Println("lb usage:")
	printCommand(ClipCommand, "entry", "copy the entry's value into the clipboard")
	printCommand(EnvCommand, "", "display environment variable information")
	printCommand(FindCommand, "criteria", "perform a simplistic text search over the entry keys")
	printCommand(HelpCommand, "", "show this usage information")
	printCommand(InsertCommand, "entry", "insert a new entry into the store")
	printSubCommand(InsertCommand, InsertMultiCommand, "entry", "insert a multi-line entry")
	printCommand(ListCommand, "", "list entries")
	printCommand(MoveCommand, "src dst", "move an entry from one location to another with the store")
	printCommand(RemoveCommand, "entry", "remove an entry from the store")
	printCommand(ShowCommand, "entry", "show the entry's value")
	printCommand(TOTPCommand, "entry", "display an updating totp generated code")
	printSubCommand(TOTPCommand, TOTPClipCommand, "entry", "copy totp code to clipboard")
	printSubCommand(TOTPCommand, TOTPListCommand, "", "list entries with totp settings")
	printSubCommand(TOTPCommand, TOTPOnceCommand, "entry", "display the first generated code")
	printSubCommand(TOTPCommand, TOTPShortCommand, "entry", "display the first generated code with no details")
	printCommand(VersionCommand, "", "display version information")
}
