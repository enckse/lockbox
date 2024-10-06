// Package app common objects
package app

import (
	"bytes"
	"embed"
	"fmt"
	"slices"
	"text/template"

	"github.com/seanenck/lockbox/internal/config"
)

type (
	// Completions handles the inputs to completions for templating
	Completions struct {
		InsertCommand       string
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
		Options             []CompletionOption
		TOTPSubCommands     []CompletionOption
		Conditionals        struct {
			Not struct {
				ReadOnly string
				NoClip   string
				NoTOTP   string
				AskMode  string
			}
		}
	}
	// CompletionOption are conditional wrapped logic for options that may be disabled
	CompletionOption struct {
		Conditional string
		Key         string
	}
	shellPreparer interface {
		ShellIsNotConditional(string) string
	}
	emptyShellPreparer struct{}
)

//go:embed shell/*
var shell embed.FS

func (e emptyShellPreparer) ShellIsNotConditional(s string) string {
	return fmt.Sprintf(config.ShellIsNotConditional, "1", s)
}

func newGenOptions(defaults []string, kv map[string]shellPreparer) []CompletionOption {
	genOption := func(to []CompletionOption, command string, prep shellPreparer, compareTo string) []CompletionOption {
		val := prep.ShellIsNotConditional(compareTo)
		return append(to, CompletionOption{val, command})
	}
	opt := []CompletionOption{}
	emptyPrepare := emptyShellPreparer{}
	for _, a := range defaults {
		opt = genOption(opt, a, emptyPrepare, "0")
	}
	for key, env := range kv {
		opt = genOption(opt, key, env, config.YesValue)
	}
	return opt
}

// GenerateCompletions handles creating shell completion outputs
func GenerateCompletions(completionType, exe string) ([]string, error) {
	if !slices.Contains(completionTypes, completionType) {
		return nil, fmt.Errorf("unknown completion request: %s", completionType)
	}
	c := Completions{
		Executable:          exe,
		InsertCommand:       InsertCommand,
		RemoveCommand:       RemoveCommand,
		TOTPListCommand:     TOTPListCommand,
		ClipCommand:         ClipCommand,
		ShowCommand:         ShowCommand,
		MultiLineCommand:    MultiLineCommand,
		JSONCommand:         JSONCommand,
		HelpCommand:         HelpCommand,
		HelpAdvancedCommand: HelpAdvancedCommand,
		TOTPCommand:         TOTPCommand,
		MoveCommand:         MoveCommand,
		DoList:              fmt.Sprintf("%s %s", exe, ListCommand),
		DoTOTPList:          fmt.Sprintf("%s %s %s", exe, TOTPCommand, TOTPListCommand),
	}
	c.Conditionals.Not.ReadOnly = config.EnvReadOnly.ShellIsNotConditional(config.YesValue)
	c.Conditionals.Not.NoClip = config.EnvNoClip.ShellIsNotConditional(config.YesValue)
	c.Conditionals.Not.NoTOTP = config.EnvNoTOTP.ShellIsNotConditional(config.YesValue)
	c.Conditionals.Not.AskMode = config.KeyModeAskConditional()

	c.Options = newGenOptions([]string{EnvCommand, HelpCommand, ListCommand, ShowCommand, VersionCommand, JSONCommand},
		map[string]shellPreparer{
			ClipCommand:             config.EnvNoClip,
			TOTPCommand:             config.EnvNoTOTP,
			MoveCommand:             config.EnvReadOnly,
			RemoveCommand:           config.EnvReadOnly,
			InsertCommand:           config.EnvReadOnly,
			MultiLineCommand:        config.EnvReadOnly,
			PasswordGenerateCommand: config.EnvNoPasswordGen,
		})
	c.TOTPSubCommands = newGenOptions([]string{TOTPMinimalCommand, TOTPOnceCommand, TOTPShowCommand},
		map[string]shellPreparer{
			TOTPClipCommand:   config.EnvNoClip,
			TOTPInsertCommand: config.EnvReadOnly,
		})
	using, err := readShell(completionType)
	if err != nil {
		return nil, err
	}
	s, err := templateScript(using, c)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil
}

func readShell(file string) (string, error) {
	return readEmbedded(fmt.Sprintf("%s.sh", file), "shell", shell)
}

func templateScript(script string, c Completions) (string, error) {
	t, err := template.New("t").Parse(script)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		return "", err
	}
	return buf.String(), nil
}
