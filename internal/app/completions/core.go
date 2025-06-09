// Package completions generations shell completions
package completions

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"text/template"

	"git.sr.ht/~enckse/lockbox/internal/app/commands"
)

type (
	// Template handles the inputs to completions for templating
	Template struct {
		InsertCommand       string
		TOTPListCommand     string
		RemoveCommand       string
		UnsetCommand        string
		ClipCommand         string
		ShowCommand         string
		MultiLineCommand    string
		MoveCommand         string
		TOTPCommand         string
		DoTOTPList          string
		DoList              string
		DoGroups            string
		Executable          string
		JSONCommand         string
		HelpCommand         string
		HelpAdvancedCommand string
		HelpConfigCommand   string
		ExportCommand       string
		Options             []string
		TOTPSubCommands     []string
	}
)

//go:embed shell/*
var shell embed.FS

// Generate handles creating shell completion outputs
func Generate(completionType, exe string) ([]string, error) {
	if !slices.Contains(commands.CompletionTypes, completionType) {
		return nil, fmt.Errorf("unknown completion request: %s", completionType)
	}
	c := Template{
		Executable:          exe,
		InsertCommand:       commands.Insert,
		UnsetCommand:        commands.Unset,
		RemoveCommand:       commands.Remove,
		TOTPListCommand:     commands.TOTPList,
		ClipCommand:         commands.Clip,
		ShowCommand:         commands.Show,
		JSONCommand:         commands.JSON,
		HelpCommand:         commands.Help,
		HelpAdvancedCommand: commands.HelpAdvanced,
		HelpConfigCommand:   commands.HelpConfig,
		TOTPCommand:         commands.TOTP,
		MoveCommand:         commands.Move,
		DoList:              fmt.Sprintf("%s %s", exe, commands.List),
		DoGroups:            fmt.Sprintf("%s %s", exe, commands.Groups),
		DoTOTPList:          fmt.Sprintf("%s %s %s", exe, commands.TOTP, commands.TOTPList),
		ExportCommand:       fmt.Sprintf("%s %s %s", exe, commands.Env, commands.Completions),
	}

	c.Options = []string{commands.Help, commands.List, commands.Show, commands.Version, commands.JSON, commands.Groups, commands.Clip, commands.TOTP, commands.Move, commands.Remove, commands.Insert, commands.Unset}
	c.TOTPSubCommands = []string{commands.TOTPMinimal, commands.TOTPOnce, commands.TOTPShow, commands.TOTPURL, commands.TOTPSeed, commands.TOTPClip}
	sort.Strings(c.Options)
	sort.Strings(c.TOTPSubCommands)

	using, err := shell.ReadFile(filepath.Join("shell", fmt.Sprintf("%s.sh", completionType)))
	if err != nil {
		return nil, err
	}
	s, err := templateScript(string(using), c)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil
}

func templateScript(script string, c Template) (string, error) {
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
