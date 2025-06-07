// Package help manages usage information
package help

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"git.sr.ht/~enckse/lockbox/internal/app/commands"
	"git.sr.ht/~enckse/lockbox/internal/backend"
	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/output"
)

const (
	docDir   = "doc"
	textFile = ".txt"
)

//go:embed doc/*
var docs embed.FS

type (
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
		Config             struct {
			Env  string
			Home string
			XDG  string
		}
		ReKey struct {
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
	const (
		isEntry  = "entry"
		isFilter = "filter"
	)
	var results []string
	results = append(results, command(commands.Clip, isEntry, "copy the entry's value into the clipboard"))
	results = append(results, command(commands.Completions, "<shell>", "generate completions via auto-detection"))
	for _, c := range commands.CompletionTypes {
		results = append(results, subCommand(commands.Completions, c, "", fmt.Sprintf("generate %s completions", c)))
	}
	results = append(results, command(commands.Env, "", "display configured variable information"))
	results = append(results, command(commands.Help, "", "show this usage information"))
	results = append(results, subCommand(commands.Help, commands.HelpAdvanced, "", "display verbose help information"))
	results = append(results, subCommand(commands.Help, commands.HelpConfig, "", "display verbose configuration information"))
	results = append(results, command(commands.Insert, isEntry, "insert a new entry into the store"))
	results = append(results, command(commands.Unset, isEntry, "clear an entry value"))
	results = append(results, command(commands.JSON, isFilter, "display detailed information"))
	results = append(results, command(commands.List, "", "list entries"))
	results = append(results, command(commands.Groups, "", "list groups"))
	results = append(results, command(commands.Find, isFilter, "find matching entries"))
	results = append(results, command(commands.Move, "src dst", "move an entry from source to destination"))
	results = append(results, command(commands.PasswordGenerate, "", "generate a password"))
	results = append(results, command(commands.ReKey, "", "rekey/reinitialize the database credentials"))
	results = append(results, command(commands.Remove, isEntry, "remove an entry from the store"))
	results = append(results, command(commands.Show, isEntry, "show the entry's value"))
	results = append(results, command(commands.TOTP, isEntry, "display an updating totp generated code"))
	results = append(results, subCommand(commands.TOTP, commands.TOTPClip, isEntry, "copy totp code to clipboard"))
	results = append(results, subCommand(commands.TOTP, commands.TOTPList, "", "list entries with totp settings"))
	results = append(results, subCommand(commands.TOTP, commands.TOTPOnce, isEntry, "display the first generated code"))
	results = append(results, subCommand(commands.TOTP, commands.TOTPMinimal, isEntry, "display one generated code (no details)"))
	results = append(results, subCommand(commands.TOTP, commands.TOTPShow, isEntry, "show the totp entry"))
	results = append(results, subCommand(commands.TOTP, commands.TOTPFind, isFilter, "find matching entries with totp settings"))
	results = append(results, command(commands.Version, "", "display version information"))
	sort.Strings(results)
	usage := []string{fmt.Sprintf("%s usage:", exe)}
	if verbose {
		results = append(results, "")
		document := Documentation{
			Executable:         filepath.Base(exe),
			MoveCommand:        commands.Move,
			RemoveCommand:      commands.Remove,
			ReKeyCommand:       commands.ReKey,
			CompletionsCommand: commands.Completions,
			HelpCommand:        commands.Help,
			HelpConfigCommand:  commands.HelpConfig,
		}
		document.Config.Env = config.ConfigEnv
		document.Config.Home = config.ConfigHome
		document.Config.XDG = config.ConfigXDG
		document.ReKey.KeyFile = setDocFlag(commands.ReKeyFlags.KeyFile)
		document.ReKey.NoKey = commands.ReKeyFlags.NoKey
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
	b, err := docs.ReadFile(filepath.Join(docDir, file))
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
	return fmt.Sprintf("%s\n%s", header, output.TextWrap(0, buf.String())), nil
}

func setDocFlag(f string) string {
	return fmt.Sprintf("-%s=", f)
}
