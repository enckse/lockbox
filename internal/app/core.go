// Package app common objects
package app

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/system"
)

const (
	// TOTPCommand is the parent of totp and by defaults generates a rotating token
	TOTPCommand = "totp"
	// ConvCommand handles text conversion of the data store
	ConvCommand = "conv"
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
	HelpAdvancedCommand = "verbose"
	// RemoveCommand removes an entry
	RemoveCommand = "rm"
	// EnvCommand shows environment information used by lockbox
	EnvCommand = "env"
	// TOTPClipCommand is the argument for copying totp codes to clipboard
	TOTPClipCommand = ClipCommand
	// TOTPMinimalCommand is the argument for getting the short version of a code
	TOTPMinimalCommand = "minimal"
	// TOTPListCommand will list the totp-enabled entries
	TOTPListCommand = ListCommand
	// TOTPOnceCommand will perform like a normal totp request but not refresh
	TOTPOnceCommand = "once"
	// EnvDefaultsCommand will display the default env variables, not those set
	EnvDefaultsCommand = "defaults"
	// BashCommand is the command to generate bash completions
	BashCommand = "bash"
	// BashDefaultsCommand will generate environment agnostic completions
	BashDefaultsCommand = "defaults"
	// ReKeyCommand will rekey the underlying database
	ReKeyCommand = "rekey"
	// MultiLineCommand handles multi-line inserts (when not piped)
	MultiLineCommand = "multiline"
	// TOTPShowCommand is for showing the TOTP token
	TOTPShowCommand = ShowCommand
	// TOTPInsertCommand is for inserting totp tokens
	TOTPInsertCommand = InsertCommand
	// JSONCommand handles JSON outputs
	JSONCommand = "json"
	// ZshCommand is the command to generate zsh completions
	ZshCommand = "zsh"
	// ZshDefaultsCommand will generate environment agnostic completions
	ZshDefaultsCommand = "defaults"
)

//go:embed doc/*
var docs embed.FS

type (
	// CommandOptions define how commands operate as an application
	CommandOptions interface {
		Confirm(string) bool
		Args() []string
		Transaction() *backend.Transaction
		Writer() io.Writer
	}

	// DefaultCommand is the default CLI app type for actual execution
	DefaultCommand struct {
		args []string
		tx   *backend.Transaction
	}
	// Completions handles the inputs to completions for templating
	Completions struct {
		Options             []string
		CanClip             bool
		CanTOTP             bool
		ReadOnly            bool
		InsertCommand       string
		TOTPSubCommands     []string
		TOTPListCommand     string
		RemoveCommand       string
		ClipCommand         string
		ShowCommand         string
		MultiLineCommand    string
		MoveCommand         string
		TOTPCommand         string
		DoTOTPList          string
		DoList              string
		Executable          string
		JSONCommand         string
		HelpCommand         string
		HelpAdvancedCommand string
	}
)

// NewDefaultCommand creates a new app command
func NewDefaultCommand(args []string) (*DefaultCommand, error) {
	t, err := backend.NewTransaction()
	if err != nil {
		return nil, err
	}
	return &DefaultCommand{args: args, tx: t}, nil
}

// Args will get the args passed to the application
func (a *DefaultCommand) Args() []string {
	return a.args
}

// Writer will get stdout
func (a *DefaultCommand) Writer() io.Writer {
	return os.Stdout
}

// Transaction will return the backend transaction
func (a *DefaultCommand) Transaction() *backend.Transaction {
	return a.tx
}

// Confirm will confirm with the user (dying if something abnormal happens)
func (a *DefaultCommand) Confirm(prompt string) bool {
	yesNo, err := system.ConfirmYesNoPrompt(prompt)
	if err != nil {
		Die(fmt.Sprintf("failed to read stdin for confirmation: %v", err))
	}
	return yesNo
}

// Die will print a message and exit (non-zero)
func Die(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

// SetArgs allow updating the command args
func (a *DefaultCommand) SetArgs(args ...string) {
	a.args = args
}

// IsPipe will indicate if we're receiving pipe input
func (a *DefaultCommand) IsPipe() bool {
	return system.IsInputFromPipe()
}

// Input will read user input
func (a *DefaultCommand) Input(pipe, multi bool) ([]byte, error) {
	return system.GetUserInputPassword(pipe, multi)
}

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

// GenerateCompletions handles creating shell completion outputs
func GenerateCompletions(isBash, defaults bool) ([]string, error) {
	name, err := exeName()
	if err != nil {
		return nil, err
	}
	c := Completions{
		Executable:          name,
		InsertCommand:       InsertCommand,
		RemoveCommand:       RemoveCommand,
		TOTPSubCommands:     []string{TOTPMinimalCommand, TOTPOnceCommand, TOTPShowCommand},
		TOTPListCommand:     TOTPListCommand,
		ClipCommand:         ClipCommand,
		ShowCommand:         ShowCommand,
		MultiLineCommand:    MultiLineCommand,
		JSONCommand:         JSONCommand,
		HelpCommand:         HelpCommand,
		HelpAdvancedCommand: HelpAdvancedCommand,
		TOTPCommand:         TOTPCommand,
		MoveCommand:         MoveCommand,
		DoList:              fmt.Sprintf("%s %s", name, ListCommand),
		DoTOTPList:          fmt.Sprintf("%s %s %s", name, TOTPCommand, TOTPListCommand),
		Options:             []string{MultiLineCommand, EnvCommand, HelpCommand, ListCommand, ShowCommand, VersionCommand, JSONCommand},
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
	if c.CanClip {
		c.Options = append(c.Options, ClipCommand)
		c.TOTPSubCommands = append(c.TOTPSubCommands, TOTPClipCommand)
	}
	if !c.ReadOnly {
		c.Options = append(c.Options, MoveCommand, RemoveCommand, InsertCommand)
		c.TOTPSubCommands = append(c.TOTPSubCommands, TOTPInsertCommand)
	}
	if c.CanTOTP {
		c.Options = append(c.Options, TOTPCommand)
	}
	using, err := readDoc("zsh")
	if err != nil {
		return nil, err
	}
	if isBash {
		using, err = readDoc("bash")
		if err != nil {
			return nil, err
		}
	}
	t, err := template.New("t").Parse(using)
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
	results = append(results, command(HelpCommand, "", "show this usage information"))
	results = append(results, subCommand(HelpCommand, HelpAdvancedCommand, "", "display verbose help information"))
	results = append(results, command(InsertCommand, "entry", "insert a new entry into the store"))
	results = append(results, command(JSONCommand, "filter", "display detailed information"))
	results = append(results, command(ListCommand, "", "list entries"))
	results = append(results, command(MoveCommand, "src dst", "move an entry from source to destination"))
	results = append(results, command(MultiLineCommand, "entry", "insert a multiline entry into the store"))
	results = append(results, command(RemoveCommand, "entry", "remove an entry from the store"))
	results = append(results, command(ShowCommand, "entry", "show the entry's value"))
	results = append(results, command(TOTPCommand, "entry", "display an updating totp generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPClipCommand, "entry", "copy totp code to clipboard"))
	results = append(results, subCommand(TOTPCommand, TOTPInsertCommand, "entry", "insert a new totp entry into the store"))
	results = append(results, subCommand(TOTPCommand, TOTPListCommand, "", "list entries with totp settings"))
	results = append(results, subCommand(TOTPCommand, TOTPOnceCommand, "entry", "display the first generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPMinimalCommand, "entry", "display the first generated code (no details)"))
	results = append(results, subCommand(TOTPCommand, TOTPShowCommand, "entry", "show the totp entry"))
	results = append(results, command(VersionCommand, "", "display version information"))
	results = append(results, command(ZshCommand, "", "generate user environment zsh completion"))
	results = append(results, subCommand(ZshCommand, ZshDefaultsCommand, "", "generate default zsh completion"))
	sort.Strings(results)
	usage := []string{fmt.Sprintf("%s usage:", name)}
	if verbose {
		results = append(results, "")
		doc, err := readDoc("details")
		if err != nil {
			return nil, err
		}
		results = append(results, strings.Split(strings.TrimSpace(doc), "\n")...)
		results = append(results, "")
		results = append(results, inputs.ListEnvironmentVariables(false)...)
	}
	return append(usage, results...), nil
}

func readDoc(doc string) (string, error) {
	b, err := docs.ReadFile(filepath.Join("doc", doc))
	if err != nil {
		return "", err
	}
	return string(b), err
}
