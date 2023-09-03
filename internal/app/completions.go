// Package app common objects
package app

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/enckse/lockbox/internal/config"
)

const (
	noList = "echo \"\""
)

type (
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

// GenerateCompletions handles creating shell completion outputs
func GenerateCompletions(isBash, defaults bool, exe string) ([]string, error) {
	c := Completions{
		Executable:          exe,
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
		DoList:              fmt.Sprintf("%s %s", exe, ListCommand),
		DoTOTPList:          fmt.Sprintf("%s %s %s", exe, TOTPCommand, TOTPListCommand),
		Options:             []string{EnvCommand, HelpCommand, ListCommand, ShowCommand, VersionCommand, JSONCommand},
	}
	isReadOnly := false
	isClip := true
	isTOTP := true
	if !defaults {
		ro, err := config.EnvReadOnly.Get()
		if err != nil {
			return nil, err
		}
		isReadOnly = ro
		noClip, err := config.EnvNoClip.Get()
		if err != nil {
			return nil, err
		}
		if noClip {
			isClip = false
		}
		noTOTP, err := config.EnvNoTOTP.Get()
		if err != nil {
			return nil, err
		}
		if noTOTP {
			isTOTP = false
		}
		k, err := config.GetKey()
		if err != nil {
			return nil, err
		}
		if k != nil && k.Interactive() {
			c.DoList = noList
			c.DoTOTPList = noList
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
		c.Options = append(c.Options, MoveCommand, RemoveCommand, InsertCommand, MultiLineCommand)
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
