// Package inputs handles user inputs/UI elements.
package inputs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/shell"
)

const (
	colorWindowDelimiter = ","
	colorWindowSpan      = ":"
	yes                  = "yes"
	no                   = "no"
	// MacOSPlatform is the macos indicator for platform
	MacOSPlatform = "macos"
	// LinuxWaylandPlatform for linux+wayland
	LinuxWaylandPlatform = "linux-wayland"
	// LinuxXPlatform for linux+X
	LinuxXPlatform = "linux-x"
	// WindowsLinuxPlatform for WSL subsystems
	WindowsLinuxPlatform = "wsl"
	unknownPlatform      = ""
)

type (
	// JSONOutputMode is the output mode definition
	JSONOutputMode    string
	environmentOutput struct {
		showValues bool
	}
	// SystemPlatform represents the platform lockbox is running on.
	SystemPlatform  string
	environmentBase struct {
		key         string
		desc        string
		requirement string
	}
	// EnvironmentInt are environment settings that are integers
	EnvironmentInt struct {
		environmentBase
		defaultValue int
		allowZero    bool
		shortDesc    string
	}
	// EnvironmentBool are environment settings that are booleans
	EnvironmentBool struct {
		environmentBase
		defaultValue bool
	}
	// EnvironmentString are string-based settings
	EnvironmentString struct {
		environmentBase
		canDefault   bool
		defaultValue string
		allowed      []string
	}
	// EnvironmentCommand are settings that are parsed as shell commands
	EnvironmentCommand struct {
		environmentBase
	}
	// EnvironmentFormatter allows for sending a string into a get request
	EnvironmentFormatter struct {
		environmentBase
		allowed string
		fxn     func(string, string) string
	}
	printer interface {
		values() (string, []string)
		self() environmentBase
	}
	// ColorWindow for handling terminal colors based on timing
	ColorWindow struct {
		Start int
		End   int
	}
)

func shlex(in string) ([]string, error) {
	return shell.Fields(in, os.Getenv)
}

func environOrDefault(envKey, defaultValue string) string {
	val := os.Getenv(envKey)
	if strings.TrimSpace(val) == "" {
		return defaultValue
	}
	return val
}

// Get will get the boolean value for the setting
func (e EnvironmentBool) Get() (bool, error) {
	read := strings.ToLower(strings.TrimSpace(os.Getenv(e.key)))
	switch read {
	case no:
		return false, nil
	case yes:
		return true, nil
	case "":
		return e.defaultValue, nil
	}

	return false, fmt.Errorf("invalid yes/no env value for %s", e.key)
}

// Get will get the integer value for the setting
func (e EnvironmentInt) Get() (int, error) {
	val := e.defaultValue
	use := os.Getenv(e.key)
	if use != "" {
		i, err := strconv.Atoi(use)
		if err != nil {
			return -1, err
		}
		invalid := false
		check := ""
		if e.allowZero {
			check = "="
		}
		switch i {
		case 0:
			invalid = !e.allowZero
		default:
			invalid = i < 0
		}
		if invalid {
			return -1, fmt.Errorf("%s must be >%s 0", e.shortDesc, check)
		}
		val = i
	}
	return val, nil
}

// Get will read the string from the environment
func (e EnvironmentString) Get() string {
	if !e.canDefault {
		return os.Getenv(e.key)
	}
	return environOrDefault(e.key, e.defaultValue)
}

// Get will read (and shlex) the value if set
func (e EnvironmentCommand) Get() ([]string, error) {
	value := environOrDefault(e.key, "")
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	return shlex(value)
}

// KeyValue will get the string representation of the key+value
func (e environmentBase) KeyValue(value string) string {
	return fmt.Sprintf("%s=%s", e.key, value)
}

// Set will do an environment set for the value to key
func (e environmentBase) Set(value string) {
	os.Setenv(e.key, value)
}

// Get will retrieve the value with the formatted input included
func (e EnvironmentFormatter) Get(value string) string {
	return e.fxn(e.key, value)
}

func (e EnvironmentString) values() (string, []string) {
	return e.defaultValue, e.allowed
}

func (e environmentBase) self() environmentBase {
	return e
}

func (e EnvironmentBool) values() (string, []string) {
	val := no
	if e.defaultValue {
		val = yes
	}
	return val, []string{yes, no}
}

func (e EnvironmentInt) values() (string, []string) {
	return fmt.Sprintf("%d", e.defaultValue), []string{"integer"}
}

func (e EnvironmentFormatter) values() (string, []string) {
	return strings.ReplaceAll(strings.ReplaceAll(EnvFormatTOTP.Get("%s"), "%25s", "%s"), "&", " \\\n           &"), []string{e.allowed}
}

func (e EnvironmentCommand) values() (string, []string) {
	return detectedValue, []string{commandArgsExample}
}

// NewPlatform gets a new system platform.
func NewPlatform() (SystemPlatform, error) {
	env := EnvPlatform.Get()
	if env != "" {
		for _, p := range EnvPlatform.allowed {
			if p == env {
				return SystemPlatform(p), nil
			}
		}
		return unknownPlatform, errors.New("unknown platform mode")
	}
	b, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return unknownPlatform, err
	}
	raw := strings.ToLower(strings.TrimSpace(string(b)))
	parts := strings.Split(raw, " ")
	switch parts[0] {
	case "darwin":
		return MacOSPlatform, nil
	case "linux":
		if strings.Contains(raw, "microsoft-standard-wsl") {
			return WindowsLinuxPlatform, nil
		}
		if strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			if strings.TrimSpace(os.Getenv("DISPLAY")) == "" {
				return unknownPlatform, errors.New("unable to detect linux clipboard mode")
			}
			return LinuxXPlatform, nil
		}
		return LinuxWaylandPlatform, nil
	}
	return unknownPlatform, errors.New("unable to detect clipboard mode")
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
