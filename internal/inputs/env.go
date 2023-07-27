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

var (
	isYesNoArgs = []string{yes, no}
	intArgs     = []string{"integer"}
)

type (
	// SystemPlatform represents the platform lockbox is running on.
	SystemPlatform  string
	environmentBase struct {
		key string
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
	}
	// EnvironmentCommand are settings that are parsed as shell commands
	EnvironmentCommand struct {
		environmentBase
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

// EnvironOrDefault will get the environment value OR default if env is not set.
func EnvironOrDefault(envKey, defaultValue string) string {
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
	return EnvironOrDefault(e.key, e.defaultValue)
}

// Get will read (and shlex) the value if set
func (e EnvironmentCommand) Get() ([]string, error) {
	value := EnvironOrDefault(e.key, "")
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	return shlex(value)
}

func (e environmentBase) Set(value string) string {
	return fmt.Sprintf("%s=%s", e.key, value)
}
