// Package inputs handles user inputs/UI elements.
package inputs

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/system"
	"mvdan.cc/sh/v3/shell"
)

const (
	prefixKey      = "LOCKBOX_"
	noClipEnv      = prefixKey + "NOCLIP"
	noColorEnv     = prefixKey + "NOCOLOR"
	interactiveEnv = prefixKey + "INTERACTIVE"
	readOnlyEnv    = prefixKey + "READONLY"
	fieldTOTPEnv   = prefixKey + "TOTP"
	clipBaseEnv    = prefixKey + "CLIP_"
	formatTOTPEnv  = fieldTOTPEnv + "_FORMAT"
	keyModeEnv     = prefixKey + "KEYMODE"
	keyEnv         = prefixKey + "KEY"
	// KeyFileEnv is an OPTIONAL keyfile for the database
	KeyFileEnv     = prefixKey + "KEYFILE"
	plainKeyMode   = "plaintext"
	commandKeyMode = "command"
	// PlatformEnv is the platform lb is running on.
	PlatformEnv = prefixKey + "PLATFORM"
	// StoreEnv is the location of the filesystem store that lb is operating on.
	StoreEnv   = prefixKey + "STORE"
	clipMaxEnv = clipBaseEnv + "MAX"
	// ColorBetweenEnv is a comma-delimited list of times to color totp outputs (e.g. 0:5,30:35 which is the default).
	ColorBetweenEnv = fieldTOTPEnv + "_BETWEEN"
	// MaxTOTPTime indicate how long TOTP tokens will be shown
	MaxTOTPTime = fieldTOTPEnv + "_MAX"
	// ClipPasteEnv allows overriding the clipboard paste command
	ClipPasteEnv = clipBaseEnv + "PASTE"
	// ClipCopyEnv allows overriding the clipboard copy command
	ClipCopyEnv        = clipBaseEnv + "COPY"
	clipOSC52Env       = clipBaseEnv + "OSC52"
	defaultTOTPField   = "totp"
	commandArgsExample = "[cmd args...]"
	// MacOSPlatform is the macos indicator for platform
	MacOSPlatform = "macos"
	// LinuxWaylandPlatform for linux+wayland
	LinuxWaylandPlatform = "linux-wayland"
	// LinuxXPlatform for linux+X
	LinuxXPlatform = "linux-x"
	// WindowsLinuxPlatform for WSL subsystems
	WindowsLinuxPlatform = "wsl"
	defaultMaxClipboard  = 45
	detectedValue        = "(detected)"
	noTOTPEnv            = prefixKey + "NOTOTP"
	// HookDirEnv represents a stored location for user hooks
	HookDirEnv = prefixKey + "HOOKDIR"
	// ModTimeEnv is modtime override ability for entries
	ModTimeEnv = prefixKey + "SET_MODTIME"
	// ModTimeFormat is the expected modtime format
	ModTimeFormat = time.RFC3339
	// MaxTOTPTimeDefault is the max TOTP time to run (default)
	MaxTOTPTimeDefault = "120"
	// JSONDataOutputEnv controls how JSON is output
	JSONDataOutputEnv = prefixKey + "JSON_DATA_OUTPUT"
	defaultHashLength = 0
	hashJSONLengthEnv = JSONDataOutputEnv + "_HASH_LENGTH"
)

var isYesNoArgs = []string{system.Yes, system.No}

type (
	environmentOutput struct {
		showValues bool
	}
	// SystemPlatform represents the platform lockbox is running on.
	SystemPlatform string
)

// GetReKey will get the rekey environment settings
func GetReKey(args []string) ([]string, error) {
	set := flag.NewFlagSet("rekey", flag.ExitOnError)
	store := set.String("store", "", "new store")
	key := set.String("key", "", "new key")
	keyFile := set.String("keyfile", "", "new keyfile")
	keyMode := set.String("keymode", "", "new keymode")
	if err := set.Parse(args); err != nil {
		return nil, err
	}
	mapped := map[string]string{
		keyModeEnv: *keyMode,
		keyEnv:     *key,
		KeyFileEnv: *keyFile,
		StoreEnv:   *store,
	}
	hasStore := false
	hasKey := false
	hasKeyFile := false
	var out []string
	for k, val := range mapped {
		if val != "" {
			switch k {
			case StoreEnv:
				hasStore = true
			case keyEnv:
				hasKey = true
			case KeyFileEnv:
				hasKeyFile = true
			}
		}
		out = append(out, fmt.Sprintf("%s=%s", k, val))
	}
	sort.Strings(out)
	if !hasStore || (!hasKey && !hasKeyFile) {
		return nil, fmt.Errorf("missing required arguments for rekey: %s", strings.Join(out, " "))
	}
	return out, nil
}

// GetClipboardMax will get max time to keep an entry in the clipboard before clearing
func GetClipboardMax() (int, error) {
	return getPositiveIntEnv(defaultMaxClipboard, clipMaxEnv, "clipboard max time", false)
}

// GetHashLength will get the maximum hash length allowed in JSON output hashing mode
func GetHashLength() (int, error) {
	return getPositiveIntEnv(defaultHashLength, hashJSONLengthEnv, "hash length", true)
}

func getPositiveIntEnv(defaultVal int, key, desc string, canBeZero bool) (int, error) {
	val := defaultVal
	use := os.Getenv(key)
	if use != "" {
		i, err := strconv.Atoi(use)
		if err != nil {
			return -1, err
		}
		invalid := false
		check := ""
		if canBeZero {
			check = "="
		}
		switch i {
		case 0:
			invalid = !canBeZero
		default:
			invalid = i < 0
		}
		if invalid {
			return -1, fmt.Errorf("%s must be >%s 0", desc, check)
		}
		val = i
	}
	return val, nil
}

// GetKey will get the encryption key setup for lb
func GetKey() ([]byte, error) {
	useKeyMode := os.Getenv(keyModeEnv)
	if useKeyMode == "" {
		useKeyMode = commandKeyMode
	}
	useKey := os.Getenv(keyEnv)
	if useKey == "" {
		return nil, errors.New("no key given")
	}
	b, err := getKey(useKeyMode, useKey)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, errors.New("key is empty")
	}
	return b, nil
}

// Shlex will do simple shell command lex-ing
func Shlex(in string) ([]string, error) {
	return shell.Fields(in, os.Getenv)
}

func getKey(keyMode, name string) ([]byte, error) {
	var data []byte
	switch keyMode {
	case commandKeyMode:
		parts, err := Shlex(name)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		b, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		data = b
	case plainKeyMode:
		data = []byte(name)
	default:
		return nil, errors.New("unknown keymode")
	}
	return []byte(strings.TrimSpace(string(data))), nil
}

func isYesNoEnv(defaultValue bool, envKey string) (bool, error) {
	read := system.EnvironValue(envKey)
	switch read {
	case system.NoValue:
		return false, nil
	case system.YesValue:
		return true, nil
	case system.EmptyValue:
		return defaultValue, nil
	}

	return false, fmt.Errorf("invalid yes/no env value for %s", envKey)
}

// IsClipOSC52 indicates if OSC52 mode is enabled
func IsClipOSC52() (bool, error) {
	return isYesNoEnv(false, clipOSC52Env)
}

// IsNoTOTP indicates if TOTP is disabled
func IsNoTOTP() (bool, error) {
	return isYesNoEnv(false, noTOTPEnv)
}

// IsReadOnly indicates to operate in readonly, no writing to file allowed
func IsReadOnly() (bool, error) {
	return isYesNoEnv(false, readOnlyEnv)
}

// IsNoClipEnabled indicates if clipboard mode is enabled.
func IsNoClipEnabled() (bool, error) {
	return isYesNoEnv(false, noClipEnv)
}

// IsNoColorEnabled indicates if the flag is set to disable color.
func IsNoColorEnabled() (bool, error) {
	return isYesNoEnv(false, noColorEnv)
}

// IsInteractive indicates if running as a user UI experience.
func IsInteractive() (bool, error) {
	return isYesNoEnv(true, interactiveEnv)
}

// TOTPToken gets the name of the totp special case tokens
func TOTPToken() string {
	return system.EnvironOrDefault(fieldTOTPEnv, defaultTOTPField)
}

func (o environmentOutput) formatEnvironmentVariable(required bool, name, val, desc string, allowed []string) string {
	value := val
	if o.showValues {
		value = os.Getenv(name)
	}
	if len(value) == 0 {
		value = "(unset)"
	}
	description := strings.ReplaceAll(desc, "\n", "\n  ")
	return fmt.Sprintf("\n%s\n  %s\n\n  required: %t\n  value: %s\n  options: %s\n", name, description, required, value, strings.Join(allowed, "|"))
}

// PlatformSet returns the list of possible platforms
func PlatformSet() []string {
	return []string{
		MacOSPlatform,
		LinuxWaylandPlatform,
		LinuxXPlatform,
		WindowsLinuxPlatform,
	}
}

// ListEnvironmentVariables will print information about env variables and potential/set values
func ListEnvironmentVariables(showValues bool) []string {
	e := environmentOutput{showValues: showValues}
	var results []string
	results = append(results, e.formatEnvironmentVariable(true, StoreEnv, "", "directory to the database file", []string{"file"}))
	results = append(results, e.formatEnvironmentVariable(true, keyModeEnv, commandKeyMode, "how to retrieve the database store password", []string{commandKeyMode, plainKeyMode}))
	results = append(results, e.formatEnvironmentVariable(true, keyEnv, "", fmt.Sprintf("the database key ('%s' mode) or command to run ('%s' mode)\nto retrieve the database password", plainKeyMode, commandKeyMode), []string{commandArgsExample, "password"}))
	results = append(results, e.formatEnvironmentVariable(false, noClipEnv, system.No, "disable clipboard operations", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, noColorEnv, system.No, "disable terminal colors", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, interactiveEnv, system.Yes, "enable interactive mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, readOnlyEnv, system.No, "operate in readonly mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, fieldTOTPEnv, defaultTOTPField, "attribute name to store TOTP tokens within the database", []string{"string"}))
	results = append(results, e.formatEnvironmentVariable(false, formatTOTPEnv, strings.ReplaceAll(strings.ReplaceAll(FormatTOTP("%s"), "%25s", "%s"), "&", " \\\n           &"), "override the otpauth url used to store totp tokens. It must have ONE format\nstring ('%s') to insert the totp base code", []string{"otpauth//url/%s/args..."}))
	results = append(results, e.formatEnvironmentVariable(false, MaxTOTPTime, MaxTOTPTimeDefault, "time, in seconds, in which to show a TOTP token before automatically exiting", []string{"integer"}))
	results = append(results, e.formatEnvironmentVariable(false, ColorBetweenEnv, TOTPDefaultBetween, "override when to set totp generated outputs to different colors, must be a\nlist of one (or more) rules where a semicolon delimits the start and end\nsecond (0-60 for each)", []string{"start:end,start:end,start:end..."}))
	results = append(results, e.formatEnvironmentVariable(false, ClipPasteEnv, detectedValue, "override the detected platform paste command", []string{commandArgsExample}))
	results = append(results, e.formatEnvironmentVariable(false, ClipCopyEnv, detectedValue, "override the detected platform copy command", []string{commandArgsExample}))
	results = append(results, e.formatEnvironmentVariable(false, clipMaxEnv, fmt.Sprintf("%d", defaultMaxClipboard), "override the amount of time before totp clears the clipboard (e.g. 10),\nmust be an integer", []string{"integer"}))
	results = append(results, e.formatEnvironmentVariable(false, PlatformEnv, detectedValue, "override the detected platform", PlatformSet()))
	results = append(results, e.formatEnvironmentVariable(false, noTOTPEnv, system.No, "disable TOTP integrations", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, HookDirEnv, "", "the path to hooks to execute on actions against the database", []string{"directory"}))
	results = append(results, e.formatEnvironmentVariable(false, clipOSC52Env, system.No, "enable OSC52 clipboard mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, KeyFileEnv, "", "additional keyfile to access/protect the database", []string{"keyfile"}))
	results = append(results, e.formatEnvironmentVariable(false, ModTimeEnv, ModTimeFormat, fmt.Sprintf("input modification time to set for the entry\n(expected format: %s)", ModTimeFormat), []string{"modtime"}))
	results = append(results, e.formatEnvironmentVariable(false, JSONDataOutputEnv, string(JSONDataOutputHash), fmt.Sprintf("changes what the data field in JSON outputs will contain\nuse '%s' with CAUTION", JSONDataOutputRaw), []string{string(JSONDataOutputRaw), string(JSONDataOutputHash), string(JSONDataOutputBlank)}))
	results = append(results, e.formatEnvironmentVariable(false, hashJSONLengthEnv, fmt.Sprintf("%d", defaultHashLength), fmt.Sprintf("maximum hash length the JSON output should contain\nwhen '%s' mode is set for JSON output", JSONDataOutputHash), []string{"integer"}))
	return results
}
