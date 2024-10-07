// Package app common objects
package app

import (
	"bytes"
	"embed"
	"fmt"
	"slices"
	"sort"
	"text/template"

	"github.com/seanenck/lockbox/internal/config"
)

const (
	shellIsNotText = `[ "%s" != "%s" ]`
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
				ReadOnly      string
				NoClip        string
				NoTOTP        string
				AskMode       string
				Ever          string
				NoPasswordGen string
			}
		}
	}
	// CompletionOption are conditional wrapped logic for options that may be disabled
	CompletionOption struct {
		Conditional string
		Key         string
	}
)

//go:embed shell/*
var shell embed.FS

func newShellIsNotEqualConditional(keyed interface{ Key() string }, right string) string {
	return fmt.Sprintf(shellIsNotText, fmt.Sprintf("$%s", keyed.Key()), right)
}

func (c Completions) newGenOptions(defaults []string, kv map[string]string) []CompletionOption {
	opt := []CompletionOption{}
	for _, a := range defaults {
		opt = append(opt, CompletionOption{c.Conditionals.Not.Ever, a})
	}
	var keys []string
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		check := kv[key]
		opt = append(opt, CompletionOption{check, key})
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
	c.Conditionals.Not.ReadOnly = newShellIsNotEqualConditional(config.EnvReadOnly, config.YesValue)
	c.Conditionals.Not.NoClip = newShellIsNotEqualConditional(config.EnvNoClip, config.YesValue)
	c.Conditionals.Not.NoTOTP = newShellIsNotEqualConditional(config.EnvNoTOTP, config.YesValue)
	c.Conditionals.Not.AskMode = newShellIsNotEqualConditional(config.EnvKeyMode, string(config.AskKeyMode))
	c.Conditionals.Not.NoPasswordGen = newShellIsNotEqualConditional(config.EnvNoPasswordGen, config.YesValue)
	c.Conditionals.Not.Ever = fmt.Sprintf(shellIsNotText, "1", "0")

	c.Options = c.newGenOptions([]string{EnvCommand, HelpCommand, ListCommand, ShowCommand, VersionCommand, JSONCommand},
		map[string]string{
			ClipCommand:             c.Conditionals.Not.NoClip,
			TOTPCommand:             c.Conditionals.Not.NoTOTP,
			MoveCommand:             c.Conditionals.Not.ReadOnly,
			RemoveCommand:           c.Conditionals.Not.ReadOnly,
			InsertCommand:           c.Conditionals.Not.ReadOnly,
			MultiLineCommand:        c.Conditionals.Not.ReadOnly,
			PasswordGenerateCommand: c.Conditionals.Not.NoPasswordGen,
		})
	c.TOTPSubCommands = c.newGenOptions([]string{TOTPMinimalCommand, TOTPOnceCommand, TOTPShowCommand},
		map[string]string{
			TOTPClipCommand:   c.Conditionals.Not.NoClip,
			TOTPInsertCommand: c.Conditionals.Not.ReadOnly,
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
