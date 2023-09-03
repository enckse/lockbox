// Package app common objects
package app

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/enckse/lockbox/internal/config"
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
		Profiles            []Profile
		DefaultProfile      Profile
		Shell               string
		CompletionEnv       string
	}

	// Profile is a completion profile
	Profile struct {
		Name      string
		CanClip   bool
		CanTOTP   bool
		CanList   bool
		ReadOnly  bool
		IsDefault bool
		env       []string
	}
)

// Env will get the environment settable value to use this profile
func (p Profile) Env() string {
	return fmt.Sprintf("%s=%s", config.EnvironmentCompletionKey, p.Display())
}

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

// Display is the profile display name
func (p Profile) Display() string {
	return strings.Join(strings.Split(strings.ToUpper(p.Name), "-")[1:], "-")
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

func loadProfiles(exe string, canFilter bool) []Profile {
	profiles := config.LoadCompletionProfiles()
	filter := config.EnvCompletion.Get()
	hasFilter := filter != "" && canFilter
	var res []Profile
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
		n.env = p.Env
		if hasFilter {
			skipped := false
			if p.Default {
				skipped = filter != "DEFAULT"
			} else {
				if filter != n.Display() {
					skipped = true
				}
			}
			if skipped {
				continue
			}
		}
		res = append(res, n)
	}
	return res
}

// GenerateCompletions handles creating shell completion outputs
func GenerateCompletions(isBash, isHelp bool, exe string) ([]string, error) {
	if isHelp {
		var h []string
		for _, p := range loadProfiles(exe, false) {
			if p.IsDefault {
				continue
			}
			text := fmt.Sprintf("export %s\n  - filtered completions\n  - useful when:\n", p.Env())
			for idx, e := range p.env {
				if idx > 0 {
					text = fmt.Sprintf("%s      and\n", text)
				}
				text = fmt.Sprintf("%s    %s\n", text, e)
			}
			h = append(h, text)
		}
		h = append(h, strings.TrimSpace(fmt.Sprintf(`
%s is not set
unset %s
export %s=<unknown>
  - default completions
`, config.EnvironmentCompletionKey, config.EnvironmentCompletionKey, config.EnvironmentCompletionKey)))
		return h, nil
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
		CompletionEnv:       fmt.Sprintf("$%s", config.EnvironmentCompletionKey),
	}

	using, err := readDoc("zsh.sh")
	if err != nil {
		return nil, err
	}
	if isBash {
		using, err = readDoc("bash.sh")
		if err != nil {
			return nil, err
		}
	}
	shellScript, err := readDoc("shell.sh")
	if err != nil {
		return nil, err
	}
	c.Profiles = loadProfiles(exe, true)
	switch len(c.Profiles) {
	case 0:
		return nil, errors.New("no profiles loaded, invalid environment setting?")
	case 1:
		c.DefaultProfile = c.Profiles[0]
	default:
		for _, p := range c.Profiles {
			if p.IsDefault {
				c.DefaultProfile = p
				break
			}
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
