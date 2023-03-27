// Package cli handles CLI helpers/commands
package cli

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/enckse/lockbox/internal/inputs"
)

const (
	// StatsCommand will display additional entry stat information
	StatsCommand = "stats"
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
	// HelpAdvancedCommand shows advanced help
	HelpAdvancedCommand = "-verbose"
	// RemoveCommand removes an entry
	RemoveCommand = "rm"
	// EnvCommand shows environment information used by lockbox
	EnvCommand = "env"
	// InsertMultiCommand handles multi-line inserts
	InsertMultiCommand = "-multi"
	// InsertTOTPCommand is a helper for totp inserts
	InsertTOTPCommand = "-totp"
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
	// ReKeyCommand will rekey the underlying database
	ReKeyCommand = "rekey"
)

var (
	//go:embed "completions.bash"
	bashCompletions string

	//go:embed "doc.txt"
	docSection string
)

type (
	// Completions handles the inputs to completions for templating
	Completions struct {
		Options             []string
		CanClip             bool
		CanTOTP             bool
		ReadOnly            bool
		InsertCommand       string
		InsertSubCommands   []string
		TOTPSubCommands     []string
		TOTPListCommand     string
		RemoveCommand       string
		ClipCommand         string
		ShowCommand         string
		MoveCommand         string
		TOTPCommand         string
		DoTOTPList          string
		DoList              string
		Executable          string
		StatsCommand        string
		HelpCommand         string
		HelpAdvancedCommand string
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

func exeName() (string, error) {
	n, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Base(n), nil
}

// BashCompletions handles creating bash completion outputs
func BashCompletions(defaults bool) ([]string, error) {
	name, err := exeName()
	if err != nil {
		return nil, err
	}
	c := Completions{
		Executable:          name,
		InsertCommand:       InsertCommand,
		RemoveCommand:       RemoveCommand,
		TOTPSubCommands:     []string{TOTPShortCommand, TOTPOnceCommand},
		TOTPListCommand:     TOTPListCommand,
		ClipCommand:         ClipCommand,
		ShowCommand:         ShowCommand,
		StatsCommand:        StatsCommand,
		InsertSubCommands:   []string{InsertMultiCommand, InsertTOTPCommand},
		HelpCommand:         HelpCommand,
		HelpAdvancedCommand: HelpAdvancedCommand,
		TOTPCommand:         TOTPCommand,
		MoveCommand:         MoveCommand,
		DoList:              fmt.Sprintf("%s %s", name, ListCommand),
		DoTOTPList:          fmt.Sprintf("%s %s %s", name, TOTPCommand, TOTPListCommand),
	}
	isReadOnly := false
	isClip := true
	isTOTP := true
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
		noTOTP, err := inputs.IsNoTOTP()
		if err != nil {
			return nil, err
		}
		if noTOTP {
			isTOTP = false
		}
	}
	c.CanClip = isClip
	c.ReadOnly = isReadOnly
	c.CanTOTP = isTOTP
	options := []string{EnvCommand, FindCommand, HelpCommand, ListCommand, ShowCommand, VersionCommand, StatsCommand}
	if c.CanClip {
		options = append(options, ClipCommand)
		c.TOTPSubCommands = append(c.TOTPSubCommands, TOTPClipCommand)
	}
	if !c.ReadOnly {
		options = append(options, MoveCommand, RemoveCommand, InsertCommand)
	}
	if c.CanTOTP {
		options = append(options, TOTPCommand)
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
func Usage(verbose bool) ([]string, error) {
	name, err := exeName()
	if err != nil {
		return nil, err
	}
	var results []string
	results = append(results, command(BashCommand, "", "generate user environment bash completion"))
	results = append(results, subCommand(BashCommand, BashDefaultsCommand, "", "generate default bash completion"))
	results = append(results, command(ClipCommand, "entry", "copy the entry's value into the clipboard"))
	results = append(results, command(EnvCommand, "", "display environment variable information"))
	results = append(results, command(FindCommand, "criteria", "perform a text search over the entry keys"))
	results = append(results, command(HelpCommand, "", "show this usage information"))
	results = append(results, subCommand(HelpCommand, HelpAdvancedCommand, "", "display verbose help information"))
	results = append(results, command(InsertCommand, "entry", "insert a new entry into the store"))
	results = append(results, subCommand(InsertCommand, InsertMultiCommand, "entry", "insert a multi-line entry"))
	results = append(results, subCommand(InsertCommand, InsertTOTPCommand, "entry", "insert a new totp entry"))
	results = append(results, command(ListCommand, "", "list entries"))
	results = append(results, command(MoveCommand, "src dst", "move an entry from source to destination"))
	results = append(results, command(RemoveCommand, "entry", "remove an entry from the store"))
	results = append(results, command(ShowCommand, "entry", "show the entry's value"))
	results = append(results, command(StatsCommand, "entry", "display entry detail information"))
	results = append(results, command(TOTPCommand, "entry", "display an updating totp generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPClipCommand, "entry", "copy totp code to clipboard"))
	results = append(results, subCommand(TOTPCommand, TOTPListCommand, "", "list entries with totp settings"))
	results = append(results, subCommand(TOTPCommand, TOTPOnceCommand, "entry", "display the first generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPShortCommand, "entry", "display the first generated code (no details)"))
	results = append(results, command(VersionCommand, "", "display version information"))
	sort.Strings(results)
	usage := []string{fmt.Sprintf("%s usage:", name)}
	if verbose {
		results = append(results, "")
		results = append(results, strings.Split(strings.TrimSpace(docSection), "\n")...)
		results = append(results, "")
		results = append(results, inputs.ListEnvironmentVariables(false)...)
	}
	return append(usage, results...), nil
}
