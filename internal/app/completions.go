// Package app common objects
package app

import (
	"bytes"
	"embed"
	"fmt"
	"slices"
	"strings"
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
		DefaultCompletion   string
		HelpCommand         string
		HelpAdvancedCommand string
		Profiles            []Profile
		Shell               string
		CompletionEnv       string
		IsYes               string
		DefaultProfile      Profile
		IsFish              bool
	}

	// Profile is a completion profile
	Profile struct {
		Name        string
		CanClip     bool
		CanTOTP     bool
		CanList     bool
		ReadOnly    bool
		IsDefault   bool
		Conditional string
	}
)

//go:embed shell/*
var shell embed.FS

// Options will list the profile options
func (p Profile) Options() []string {
	opts := []string{EnvCommand, HelpCommand, ListCommand, ShowCommand, VersionCommand, JSONCommand}
	if p.CanClip {
		opts = append(opts, ClipCommand)
	}
	if !p.ReadOnly {
		opts = append(opts, MoveCommand, RemoveCommand, InsertCommand, MultiLineCommand)
	}
	if p.CanTOTP {
		opts = append(opts, TOTPCommand)
	}
	return opts
}

// TOTPSubCommands are the list of sub commands for TOTP within the profile
func (p Profile) TOTPSubCommands() []string {
	totp := []string{TOTPMinimalCommand, TOTPOnceCommand, TOTPShowCommand}
	if p.CanClip {
		totp = append(totp, TOTPClipCommand)
	}
	if !p.ReadOnly {
		totp = append(totp, TOTPInsertCommand)
	}
	return totp
}

func loadProfiles(exe string) []Profile {
	profiles := config.LoadCompletionProfiles()
	conditionals := make(map[int][]Profile)
	maxCount := 0
	for _, p := range profiles {
		name := p.Name
		if p.Default {
			name = "default"
		}
		n := Profile{Name: fmt.Sprintf("_%s-%s", exe, name)}
		n.CanClip = p.Clip
		n.CanList = p.List
		n.CanTOTP = p.TOTP
		n.ReadOnly = !p.Write
		n.IsDefault = p.Default
		var sub []string
		for _, e := range p.Env {
			sub = append(sub, fmt.Sprintf("[ %s ]", e))
		}
		n.Conditional = strings.Join(sub, " && ")
		count := len(p.Env)
		val := conditionals[count]
		conditionals[count] = append(val, n)
		if count > maxCount {
			maxCount = count
		}
	}
	var res []Profile
	for maxCount >= 0 {
		val, ok := conditionals[maxCount]
		if ok {
			res = append(res, val...)
		}
		maxCount--
	}
	return res
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
		DefaultCompletion:   fmt.Sprintf("$%s", config.EnvDefaultCompletionKey),
		IsYes:               config.YesValue,
		IsFish:              completionType == CompletionsFishCommand,
	}

	using, err := readShell(completionType)
	if err != nil {
		return nil, err
	}
	shellScript, err := readShell("shell")
	if err != nil {
		return nil, err
	}
	c.Profiles = loadProfiles(exe)
	for _, p := range c.Profiles {
		if p.IsDefault {
			c.DefaultProfile = p
			break
		}
	}
	shell, err := templateScript(shellScript, c)
	if err != nil {
		return nil, err
	}
	c.Shell = shell
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
