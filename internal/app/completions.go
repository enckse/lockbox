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
			ReadOnly string
			NoClip   string
			NoTOTP   string
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

func newConditional(left, right string) string {
	return fmt.Sprintf("[ \"%s\" != \"%s\" ]", left, right)
}

func genOption(to []CompletionOption, command, left, right string) []CompletionOption {
	conditional := newConditional(left, right)
	return append(to, CompletionOption{conditional, command})
}

func newGenOptions(defaults ...string) []CompletionOption {
	opt := []CompletionOption{}
	for _, a := range defaults {
		opt = genOption(opt, a, "1", "0")
	}
	return opt
}

func genOptionKeyValues(to []CompletionOption, kv map[string]string) []CompletionOption {
	for key, env := range kv {
		to = genOption(to, key, fmt.Sprintf("$%s", env), config.YesValue)
	}
	return to
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
	c.Conditionals.ReadOnly = newConditional(config.EnvReadOnly.Key(), config.YesValue)
	c.Conditionals.NoClip = newConditional(config.EnvNoClip.Key(), config.YesValue)
	c.Conditionals.NoTOTP = newConditional(config.EnvNoTOTP.Key(), config.YesValue)

	cmds := newGenOptions(
		EnvCommand,
		HelpCommand,
		ListCommand,
		ShowCommand,
		VersionCommand,
		JSONCommand,
	)
	cmds = genOptionKeyValues(cmds, map[string]string{
		ClipCommand:             config.EnvNoClip.Key(),
		TOTPCommand:             config.EnvNoTOTP.Key(),
		MoveCommand:             config.EnvReadOnly.Key(),
		RemoveCommand:           config.EnvReadOnly.Key(),
		InsertCommand:           config.EnvReadOnly.Key(),
		MultiLineCommand:        config.EnvReadOnly.Key(),
		PasswordGenerateCommand: config.EnvNoPasswordGen.Key(),
	})
	c.Options = cmds
	totp := newGenOptions(TOTPMinimalCommand, TOTPOnceCommand, TOTPShowCommand)
	totp = genOptionKeyValues(totp, map[string]string{
		TOTPClipCommand:   config.EnvNoClip.Key(),
		TOTPInsertCommand: config.EnvReadOnly.Key(),
	})
	c.TOTPSubCommands = totp
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
