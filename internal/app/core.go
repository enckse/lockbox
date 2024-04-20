// Package app common objects
package app

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/enckse/lockbox/internal/backend"
	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/platform"
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
	// BashCommand is the command to generate bash completions
	BashCommand = "bash"
	// HelpShellCommand is the help output about shell variables
	HelpShellCommand = "shell"
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
	// FishCommand is the command to generate fish completions
	FishCommand = "fish"
	docDir      = "doc"
	textFile    = ".txt"
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
	// Documentation is how documentation segments are templated
	Documentation struct {
		Executable       string
		MoveCommand      string
		RemoveCommand    string
		ReKeyCommand     string
		HelpCommand      string
		HelpShellCommand string
		ReKey            struct {
			Store   string
			KeyFile string
			Key     string
			KeyMode string
		}
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
	yesNo, err := platform.ConfirmYesNoPrompt(prompt)
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
	return platform.IsInputFromPipe()
}

// Input will read user input
func (a *DefaultCommand) Input(pipe, multi bool) ([]byte, error) {
	return platform.GetUserInputPassword(pipe, multi)
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

// Usage return usage information
func Usage(verbose bool, exe string) ([]string, error) {
	var results []string
	results = append(results, command(BashCommand, "", "generate user environment bash completion"))
	results = append(results, command(ClipCommand, "entry", "copy the entry's value into the clipboard"))
	results = append(results, command(EnvCommand, "", "display environment variable information"))
	results = append(results, command(HelpCommand, "", "show this usage information"))
	results = append(results, subCommand(HelpCommand, HelpAdvancedCommand, "", "display verbose help information"))
	results = append(results, subCommand(HelpCommand, HelpShellCommand, "", "display shell variable help information"))
	results = append(results, command(InsertCommand, "entry", "insert a new entry into the store"))
	results = append(results, command(JSONCommand, "filter", "display detailed information"))
	results = append(results, command(ListCommand, "", "list entries"))
	results = append(results, command(MoveCommand, "src dst", "move an entry from source to destination"))
	results = append(results, command(MultiLineCommand, "entry", "insert a multiline entry into the store"))
	results = append(results, command(ReKeyCommand, "", "rekey/reinitialize the database credentials"))
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
	results = append(results, command(FishCommand, "", "generate user environment fish completion"))
	sort.Strings(results)
	usage := []string{fmt.Sprintf("%s usage:", exe)}
	if verbose {
		results = append(results, "")
		document := Documentation{
			Executable:       filepath.Base(exe),
			MoveCommand:      MoveCommand,
			RemoveCommand:    RemoveCommand,
			ReKeyCommand:     ReKeyCommand,
			HelpShellCommand: HelpShellCommand,
			HelpCommand:      HelpCommand,
		}
		document.ReKey.Store = setDocFlag(config.ReKeyStoreFlag)
		document.ReKey.Key = setDocFlag(config.ReKeyKeyFlag)
		document.ReKey.KeyMode = setDocFlag(config.ReKeyKeyModeFlag)
		document.ReKey.KeyFile = setDocFlag(config.ReKeyKeyModeFlag)
		files, err := docs.ReadDir(docDir)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		var env string
		for _, f := range files {
			n := f.Name()
			if !strings.HasSuffix(n, textFile) {
				continue
			}
			header := fmt.Sprintf("[%s]", strings.TrimSuffix(filepath.Base(n), textFile))
			s, err := processDoc(header, n, document)
			if err != nil {
				return nil, err
			}
			if header == "[environment]" {
				env = s
			} else {
				buf.WriteString(s)
			}
		}
		if env == "" {
			return nil, errors.New("no environment header configured")
		}
		buf.WriteString(env)
		results = append(results, strings.Split(strings.TrimSpace(buf.String()), "\n")...)
		results = append(results, "")
		results = append(results, config.ListEnvironmentVariables()...)
	}
	return append(usage, results...), nil
}

func processDoc(header, file string, doc Documentation) (string, error) {
	b, err := readDoc(file)
	if err != nil {
		return "", err
	}
	t, err := template.New("d").Parse(string(b))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, doc); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s", header, config.Wrap(0, buf.String())), nil
}

func setDocFlag(f string) string {
	return fmt.Sprintf("-%s=", f)
}

func readDoc(doc string) (string, error) {
	b, err := docs.ReadFile(filepath.Join(docDir, doc))
	if err != nil {
		return "", err
	}
	return string(b), err
}
