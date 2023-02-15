// Package inputs handles user inputs/UI elements.
package inputs

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/shlex"
)

const (
	otpAuth        = "otpauth"
	otpIssuer      = "lbissuer"
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
	// ClipPasteEnv allows overriding the clipboard paste command
	ClipPasteEnv = clipBaseEnv + "PASTE"
	// ClipCopyEnv allows overriding the clipboard copy command
	ClipCopyEnv        = clipBaseEnv + "COPY"
	clipOSC52Env       = clipBaseEnv + "OSC52"
	isYes              = "yes"
	isNo               = "no"
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
	colorWindowDelimiter = ","
	colorWindowSpan      = ":"
	detectedValue        = "(detected)"
	noTOTPEnv            = prefixKey + "NOTOTP"
	// HookDirEnv represents a stored location for user hooks
	HookDirEnv = prefixKey + "HOOKDIR"
	// ModTimeEnv is modtime override ability for entries
	ModTimeEnv = prefixKey + "SET_MODTIME"
	// ModTimeFormat is the expected modtime format
	ModTimeFormat = time.RFC3339
	reKeySuffix   = "_NEW"
)

var (
	isYesNoArgs = []string{isYes, isNo}
	// TOTPDefaultColorWindow is the default coloring rules for totp
	TOTPDefaultColorWindow = []ColorWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = toString(TOTPDefaultColorWindow)
)

type (
	environmentOutput struct {
		showValues bool
	}
	// ColorWindow for handling terminal colors based on timing
	ColorWindow struct {
		Start int
		End   int
	}
	// SystemPlatform represents the platform lockbox is running on.
	SystemPlatform string
)

// SetReKey will enable the rekeying mode for the environment
func SetReKey() error {
	hasStore := false
	hasKey := false
	hasKeyFile := false
	for _, k := range []string{keyModeEnv, keyEnv, KeyFileEnv, StoreEnv} {
		val := os.Getenv(fmt.Sprintf("%s%s", k, reKeySuffix))
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
		if err := os.Setenv(k, val); err != nil {
			return err
		}
	}
	if !hasStore || (!hasKey && !hasKeyFile) {
		return errors.New("missing required environment variables for rekey")
	}
	return nil
}

func toString(windows []ColorWindow) string {
	var results []string
	for _, w := range windows {
		results = append(results, fmt.Sprintf("%d%s%d", w.Start, colorWindowSpan, w.End))
	}
	return strings.Join(results, colorWindowDelimiter)
}

// ParseColorWindow will handle parsing a window of colors for TOTP operations
func ParseColorWindow(windowString string) ([]ColorWindow, error) {
	var rules []ColorWindow
	for _, item := range strings.Split(windowString, colorWindowDelimiter) {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		parts := strings.Split(line, colorWindowSpan)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid colorization rule found: %s", line)
		}
		s, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		e, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if s < 0 || e < 0 || e < s || s > 59 || e > 59 {
			return nil, fmt.Errorf("invalid time found for colorization rule: %s", line)
		}
		rules = append(rules, ColorWindow{Start: s, End: e})
	}
	if len(rules) == 0 {
		return nil, errors.New("invalid colorization rules for totp, none found")
	}
	return rules, nil
}

// EnvOrDefault will get the environment value OR default if env is not set.
func EnvOrDefault(envKey, defaultValue string) string {
	val := os.Getenv(envKey)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetClipboardMax will get max time to keep an entry in the clipboard before clearing
func GetClipboardMax() (int, error) {
	max := defaultMaxClipboard
	useMax := os.Getenv(clipMaxEnv)
	if useMax != "" {
		i, err := strconv.Atoi(useMax)
		if err != nil {
			return -1, err
		}
		if i < 1 {
			return -1, errors.New("clipboard max time must be greater than 0")
		}
		max = i
	}
	return max, nil
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

func getKey(keyMode, name string) ([]byte, error) {
	var data []byte
	switch keyMode {
	case commandKeyMode:
		parts, err := shlex.Split(name)
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

func isYesNoEnv(defaultValue bool, env string) (bool, error) {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(env)))
	if len(value) == 0 {
		return defaultValue, nil
	}
	switch value {
	case isNo:
		return false, nil
	case isYes:
		return true, nil
	}
	return false, fmt.Errorf("invalid yes/no env value for %s", env)
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
	return EnvOrDefault(fieldTOTPEnv, defaultTOTPField)
}

// FormatTOTP will format a totp otpauth url
func FormatTOTP(value string) string {
	if strings.HasPrefix(value, otpAuth) {
		return value
	}
	override := EnvOrDefault(formatTOTPEnv, "")
	if override != "" {
		return fmt.Sprintf(override, value)
	}
	v := url.Values{}
	v.Set("secret", value)
	v.Set("issuer", otpIssuer)
	v.Set("period", "30")
	v.Set("algorithm", "SHA1")
	v.Set("digits", "6")
	u := url.URL{
		Scheme:   otpAuth,
		Host:     "totp",
		Path:     "/" + otpIssuer + ":" + "lbaccount",
		RawQuery: v.Encode(),
	}
	return u.String()
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
	results = append(results, e.formatEnvironmentVariable(true, keyEnv, "", fmt.Sprintf("the database key ('%s' mode) or shell command to run ('%s' mode)\nto retrieve the database password", plainKeyMode, commandKeyMode), []string{commandArgsExample, "password"}))
	results = append(results, e.formatEnvironmentVariable(false, noClipEnv, isNo, "disable clipboard operations", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, noColorEnv, isNo, "disable terminal colors", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, interactiveEnv, isYes, "enable interactive mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, readOnlyEnv, isNo, "operate in readonly mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, fieldTOTPEnv, defaultTOTPField, "attribute name to store TOTP tokens within the database", []string{"string"}))
	results = append(results, e.formatEnvironmentVariable(false, formatTOTPEnv, strings.ReplaceAll(strings.ReplaceAll(FormatTOTP("%s"), "%25s", "%s"), "&", " \\\n           &"), "override the otpauth url used to store totp tokens. It must have ONE format\nstring ('%s') to insert the totp base code", []string{"otpauth//url/%s/args..."}))
	results = append(results, e.formatEnvironmentVariable(false, ColorBetweenEnv, TOTPDefaultBetween, "override when to set totp generated outputs to different colors, must be a\nlist of one (or more) rules where a semicolon delimits the start and end\nsecond (0-60 for each)", []string{"start:end,start:end,start:end..."}))
	results = append(results, e.formatEnvironmentVariable(false, ClipPasteEnv, detectedValue, "override the detected platform paste command", []string{commandArgsExample}))
	results = append(results, e.formatEnvironmentVariable(false, ClipCopyEnv, detectedValue, "override the detected platform copy command", []string{commandArgsExample}))
	results = append(results, e.formatEnvironmentVariable(false, clipMaxEnv, fmt.Sprintf("%d", defaultMaxClipboard), "override the amount of time before totp clears the clipboard (e.g. 10),\nmust be an integer", []string{"integer"}))
	results = append(results, e.formatEnvironmentVariable(false, PlatformEnv, detectedValue, "override the detected platform", PlatformSet()))
	results = append(results, e.formatEnvironmentVariable(false, noTOTPEnv, isNo, "disable TOTP integrations", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, HookDirEnv, "", "the path to hooks to execute on actions against the database", []string{"directory"}))
	results = append(results, e.formatEnvironmentVariable(false, clipOSC52Env, isNo, "enable OSC52 clipboard mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, KeyFileEnv, "", "additional keyfile to access/protect the database", []string{"keyfile"}))
	results = append(results, e.formatEnvironmentVariable(false, ModTimeEnv, ModTimeFormat, fmt.Sprintf("input modification time to set for the entry\n(expected format: %s)", ModTimeFormat), []string{"modtime"}))
	return results
}
