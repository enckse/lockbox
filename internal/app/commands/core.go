// Package commands defines available commands within the app
package commands

import (
	"slices"

	"git.sr.ht/~enckse/lockbox/internal/config"
)

const (
	// TOTP is the parent of totp and by defaults generates a rotating token
	TOTP = "totp"
	// Conv handles text conversion of the data store
	Conv = "conv"
	// ClipManager is a callback to manage clipboard clearing
	ClipManager = "clipmanager"
	// Clip will copy values to the clipboard
	Clip = "clip"
	// Insert adds a value
	Insert = "insert"
	// List lists all entries
	List = "ls"
	// Move will move source to destination
	Move = "mv"
	// Show will show the value in an entry
	Show = "show"
	// Version displays version information
	Version = "version"
	// Help shows usage
	Help = "help"
	// HelpAdvanced shows advanced help
	HelpAdvanced = "verbose"
	// HelpConfig shows configuration information
	HelpConfig = "config"
	// Remove removes an entry
	Remove = "rm"
	// Env shows environment information used by lockbox
	Env = "vars"
	// TOTPClip is the argument for copying totp codes to clipboard
	TOTPClip = Clip
	// TOTPMinimal is the argument for getting the short version of a code
	TOTPMinimal = "minimal"
	// TOTPList will list the totp-enabled entries
	TOTPList = List
	// TOTPOnce will perform like a normal totp request but not refresh
	TOTPOnce = "once"
	// CompletionsBash is the command to generate bash completions
	CompletionsBash = "bash"
	// Completions are used to generate shell completions
	Completions = "completions"
	// ReKey will rekey the underlying database
	ReKey = "rekey"
	// TOTPShow is for showing the TOTP token
	TOTPShow = Show
	// JSON handles JSON outputs
	JSON = "json"
	// CompletionsZsh is the command to generate zsh completions
	CompletionsZsh = "zsh"
	// Executable is the name of the executable
	Executable = "lb"
	// Unset indicates a value should be unset (removed) from an entity
	Unset = "unset"
	// Groups handles getting a list of groups
	Groups = "groups"
	// TOTPURL will display TOTP URL information
	TOTPURL = "url"
	// TOTPSeed will display the seed for the TOTP tokens
	TOTPSeed = "seed"
)

var (
	// CompletionTypes are shell completions that are known
	CompletionTypes = []string{CompletionsBash, CompletionsZsh}
	// ReKeyFlags are the flags used for re-keying
	ReKeyFlags = struct {
		KeyFile string
		NoKey   string
	}{"keyfile", "nokey"}
)

// AllowedInReadOnly indicates any commands that are allowed in readonly mode
func AllowedInReadOnly(cmds ...string) []string {
	if config.EnvReadOnly.Get() {
		var allowed []string
		for _, item := range cmds {
			if slices.Contains([]string{Move, Insert, Unset, Remove, ReKey}, item) {
				continue
			}
			allowed = append(allowed, item)
		}
		return allowed
	}
	return cmds
}

// ReKeyArgs is the base definition of re-keying args
type ReKeyArgs struct {
	KeyFile string
	NoKey   bool
}
