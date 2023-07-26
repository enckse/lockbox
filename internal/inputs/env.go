// Package inputs handles user inputs/UI elements.
package inputs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/shell"
)

const (
	yes = "yes"
	no  = "no"
	// MacOSPlatform is the macos indicator for platform
	MacOSPlatform = "macos"
	// LinuxWaylandPlatform for linux+wayland
	LinuxWaylandPlatform = "linux-wayland"
	// LinuxXPlatform for linux+X
	LinuxXPlatform = "linux-x"
	// WindowsLinuxPlatform for WSL subsystems
	WindowsLinuxPlatform = "wsl"
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
		key      string
		required bool
		desc     string
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
)

func shlex(in string) ([]string, error) {
	return shell.Fields(in, os.Getenv)
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
