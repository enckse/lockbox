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

	"github.com/seanenck/lockbox/internal/backend"
	"github.com/seanenck/lockbox/internal/core"
	"github.com/seanenck/lockbox/internal/platform"
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
	// HelpConfigCommand shows configuration information
	HelpConfigCommand = "config"
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
	// CompletionsBashCommand is the command to generate bash completions
	CompletionsBashCommand = "bash"
	// CompletionsCommand are used to generate shell completions
	CompletionsCommand = "completions"
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
	// CompletionsZshCommand is the command to generate zsh completions
	CompletionsZshCommand = "zsh"
	// CompletionsFishCommand is the command to generate fish completions
	CompletionsFishCommand = "fish"
	docDir                 = "doc"
	textFile               = ".txt"
	// PasswordGenerateCommand is the command to do password generation
	PasswordGenerateCommand = "pwgen"
)

var (
	//go:embed doc/*
	docs            embed.FS
	completionTypes = []string{CompletionsBashCommand, CompletionsFishCommand, CompletionsZshCommand}
)

type (
	// CommandOptions define how commands operate as an application
	CommandOptions interface {
		Confirm(string) bool
		Args() []string
		Transaction() *backend.Transaction
		Writer() io.Writer
	}

	// UserInputOptions handle user inputs (e.g. password entry)
	UserInputOptions interface {
		CommandOptions
		IsPipe() bool
		Input(bool) ([]byte, error)
	}

	// DefaultCommand is the default CLI app type for actual execution
	DefaultCommand struct {
		args []string
		tx   *backend.Transaction
	}
	// Documentation is how documentation segments are templated
	Documentation struct {
		Executable         string
		MoveCommand        string
		RemoveCommand      string
		ReKeyCommand       string
		CompletionsCommand string
		CompletionsEnv     string
		HelpCommand        string
		HelpConfigCommand  string
		ReKey              struct {
			KeyFile string
			NoKey   string
		}
		Hooks struct {
			Mode struct {
				Pre  string
				Post string
			}
			Action struct {
				Remove string
				Insert string
				Move   string
			}
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

// ReadLine handles a single stdin read
func (a DefaultCommand) ReadLine() (string, error) {
	return platform.Stdin(true)
}

// Password is how a keyer gets the user's password for rekey
func (a DefaultCommand) Password() (string, error) {
	return platform.ReadInteractivePassword()
}

// Input will read user input
func (a *DefaultCommand) Input(interactive bool) ([]byte, error) {
	return platform.GetUserInputPassword(interactive)
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
	return fmt.Sprintf("  %-18s %-10s    %s", name, arguments, desc)
}

// Usage return usage information
func Usage(verbose bool, exe string) ([]string, error) {
	var results []string
	results = append(results, command(ClipCommand, "entry", "copy the entry's value into the clipboard"))
	results = append(results, command(CompletionsCommand, "<shell>", "generate completions via auto-detection"))
	for _, c := range completionTypes {
		results = append(results, subCommand(CompletionsCommand, c, "", fmt.Sprintf("generate %s completions", c)))
	}
	results = append(results, command(EnvCommand, "", "display environment variable information"))
	results = append(results, command(HelpCommand, "", "show this usage information"))
	results = append(results, subCommand(HelpCommand, HelpAdvancedCommand, "", "display verbose help information"))
	results = append(results, subCommand(HelpCommand, HelpConfigCommand, "", "display verbose configuration information"))
	results = append(results, command(InsertCommand, "entry", "insert a new entry into the store"))
	results = append(results, command(JSONCommand, "filter", "display detailed information"))
	results = append(results, command(ListCommand, "", "list entries"))
	results = append(results, command(MoveCommand, "src dst", "move an entry from source to destination"))
	results = append(results, command(MultiLineCommand, "entry", "insert a multiline entry into the store"))
	results = append(results, command(PasswordGenerateCommand, "", "generate a password"))
	results = append(results, command(ReKeyCommand, "", "rekey/reinitialize the database credentials"))
	results = append(results, command(RemoveCommand, "entry", "remove an entry from the store"))
	results = append(results, command(ShowCommand, "entry", "show the entry's value"))
	results = append(results, command(TOTPCommand, "entry", "display an updating totp generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPClipCommand, "entry", "copy totp code to clipboard"))
	results = append(results, subCommand(TOTPCommand, TOTPInsertCommand, "entry", "insert a new totp entry into the store"))
	results = append(results, subCommand(TOTPCommand, TOTPListCommand, "", "list entries with totp settings"))
	results = append(results, subCommand(TOTPCommand, TOTPOnceCommand, "entry", "display the first generated code"))
	results = append(results, subCommand(TOTPCommand, TOTPMinimalCommand, "entry", "display one generated code (no details)"))
	results = append(results, subCommand(TOTPCommand, TOTPShowCommand, "entry", "show the totp entry"))
	results = append(results, command(VersionCommand, "", "display version information"))
	sort.Strings(results)
	usage := []string{fmt.Sprintf("%s usage:", exe)}
	if verbose {
		results = append(results, "")
		document := Documentation{
			Executable:         filepath.Base(exe),
			MoveCommand:        MoveCommand,
			RemoveCommand:      RemoveCommand,
			ReKeyCommand:       ReKeyCommand,
			CompletionsCommand: CompletionsCommand,
			HelpCommand:        HelpCommand,
			HelpConfigCommand:  HelpConfigCommand,
		}
		document.ReKey.KeyFile = setDocFlag(reKeyFlags.KeyFile)
		document.ReKey.NoKey = reKeyFlags.NoKey
		document.Hooks.Mode.Pre = string(backend.HookPre)
		document.Hooks.Mode.Post = string(backend.HookPost)
		document.Hooks.Action.Insert = string(backend.InsertAction)
		document.Hooks.Action.Remove = string(backend.RemoveAction)
		document.Hooks.Action.Move = string(backend.MoveAction)
		files, err := docs.ReadDir(docDir)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
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
			buf.WriteString(s)
		}
		results = append(results, strings.Split(strings.TrimSpace(buf.String()), "\n")...)
	}
	return append(usage, results...), nil
}

func processDoc(header, file string, doc Documentation) (string, error) {
	b, err := readEmbedded(file, docDir, docs)
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
	return fmt.Sprintf("%s\n%s", header, core.TextWrap(0, buf.String())), nil
}

func setDocFlag(f string) string {
	return fmt.Sprintf("-%s=", f)
}

func readEmbedded(file, dir string, e embed.FS) (string, error) {
	b, err := e.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return "", err
	}
	return string(b), err
}
