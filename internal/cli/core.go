// Package cli handles CLI helpers/commands
package cli

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"text/template"

	"github.com/enckse/lockbox/internal/inputs"
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
	// BashCommand is the command to generate bash completions
	BashCommand = "bash"
	// BashDefaultsCommand will generate environment agnostic completions
	BashDefaultsCommand = "-defaults"
)

var (
	//go:embed "completions.bash"
	bashCompletions string
)

type (
	// Completions handles the inputs to completions for templating
	Completions struct {
		Options            []string
		CanClip            bool
		ReadOnly           bool
		InsertCommand      string
		TOTPShortCommand   string
		TOTPOnceCommand    string
		TOTPClipCommand    string
		InsertMultiCommand string
		RemoveCommand      string
		ClipCommand        string
		ShowCommand        string
		MoveCommand        string
		TOTPCommand        string
		DoTOTPList         string
		DoList             string
	}
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

// BashCompletions handles creating bash completion outputs
func BashCompletions(defaults bool) ([]string, error) {
	c := Completions{
		InsertCommand:      InsertCommand,
		RemoveCommand:      RemoveCommand,
		TOTPShortCommand:   TOTPShortCommand,
		TOTPClipCommand:    TOTPClipCommand,
		TOTPOnceCommand:    TOTPOnceCommand,
		ClipCommand:        ClipCommand,
		ShowCommand:        ShowCommand,
		InsertMultiCommand: InsertMultiCommand,
		TOTPCommand:        TOTPCommand,
		MoveCommand:        MoveCommand,
		DoList:             fmt.Sprintf("lb %s", ListCommand),
		DoTOTPList:         fmt.Sprintf("lb %s %s", TOTPCommand, TOTPListCommand),
	}
	isReadOnly := false
	isClip := true
	if !defaults {
		ro, err := inputs.IsReadOnly()
		if err != nil {
			return nil, err
		}
		isReadOnly = ro
		noClip, err := inputs.IsNoClipEnabled()
		if err != nil {
			return nil, err
		}
		if noClip {
			isClip = false
		}
	}
	c.CanClip = isClip
	c.ReadOnly = isReadOnly
	options := []string{EnvCommand, FindCommand, HelpCommand, ListCommand, ShowCommand, TOTPCommand, VersionCommand}
	if c.CanClip {
		options = append(options, ClipCommand)
	}
	if !c.ReadOnly {
		options = append(options, MoveCommand, RemoveCommand, InsertCommand)
	}
	c.Options = options
	t, err := template.New("t").Parse(bashCompletions)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		return nil, err
	}
	return []string{buf.String()}, nil
}

// Usage return usage information
func Usage() []string {
	var results []string
	results = append(results, command(BashCommand, "", "generate bash completions"))
	results = append(results, subCommand(BashCommand, BashDefaultsCommand, "", "generate default bash completion, not user environment specific"))
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
	sort.Strings(results)
	usage := []string{"lb usage:"}
	return append(usage, results...)
}
