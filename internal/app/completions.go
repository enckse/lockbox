// Package app common objects
package app

import (
	"bytes"
	"fmt"
	"sort"
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
		Comment   string
		CanClip   bool
		CanTOTP   bool
		CanList   bool
		ReadOnly  bool
		IsDefault bool
	}
)

const (
	askProfile    = "ask"
	roProfile     = "readonly"
	noTOTPProfile = "nototp"
	noClipProfile = "noclip"
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
	return strings.Join(strings.Split(strings.ToUpper(p.Name), "-")[2:], "-")
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

func newProfile(exe string, keys []string) Profile {
	p := Profile{}
	p.CanClip = true
	p.CanList = true
	p.CanTOTP = true
	p.ReadOnly = false
	name := "profile-"
	sort.Strings(keys)
	var comments []string
	for _, k := range keys {
		name = fmt.Sprintf("%s%s-", name, k)
		switch k {
		case askProfile:
			comments = append(comments, "ask key mode = on")
			p.CanList = false
		case noTOTPProfile:
			comments = append(comments, "totp = off")
			p.CanTOTP = false
		case noClipProfile:
			comments = append(comments, "clipboard = off")
			p.CanClip = false
		case roProfile:
			comments = append(comments, "readonly = on")
			p.ReadOnly = true
		}
	}
	sort.Strings(comments)
	p.Name = newCompletionName(exe, strings.TrimSuffix(name, "-"))
	p.Comment = fmt.Sprintf("# - %s", strings.Join(comments, "\n# - "))
	return p
}

func generateProfiles(exe string, keys []string) map[string]Profile {
	m := make(map[string]Profile)
	if len(keys) == 0 {
		return m
	}
	p := newProfile(exe, keys)
	m[p.Name] = p
	for _, cur := range keys {
		var subset []string
		for _, key := range keys {
			if key == cur {
				continue
			}
			subset = append(subset, key)
		}

		for _, p := range generateProfiles(exe, subset) {
			m[p.Name] = p
		}
	}
	return m
}

func newCompletionName(exe, name string) string {
	return fmt.Sprintf("_%s-%s", exe, name)
}

// GenerateCompletions handles creating shell completion outputs
func GenerateCompletions(isBash bool, exe string) ([]string, error) {
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
	shellScript, err := readDoc("shell")
	if err != nil {
		return nil, err
	}
	profiles := generateProfiles(exe, []string{noClipProfile, roProfile, noTOTPProfile, askProfile})
	profileObjects := []Profile{}
	for _, v := range profiles {
		profileObjects = append(profileObjects, v)
	}
	sort.Slice(profileObjects, func(i, j int) bool {
		return strings.Compare(profileObjects[i].Name, profileObjects[j].Name) < 0
	})
	c.Profiles = append(c.Profiles, profileObjects...)
	c.DefaultProfile = Profile{IsDefault: true, CanClip: true, CanTOTP: true, CanList: true, ReadOnly: false, Name: newCompletionName(exe, "default")}
	c.Profiles = append(c.Profiles, c.DefaultProfile)
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
