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

var isYesNoArgs = []string{yes, no}

type (
	// SystemPlatform represents the platform lockbox is running on.
	SystemPlatform string
)

// Shlex will do simple shell command lex-ing
func Shlex(in string) ([]string, error) {
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

func isYesNoEnv(defaultValue bool, envKey string) (bool, error) {
	read := strings.ToLower(strings.TrimSpace(os.Getenv(envKey)))
	switch read {
	case no:
		return false, nil
	case yes:
		return true, nil
	case "":
		return defaultValue, nil
	}

	return false, fmt.Errorf("invalid yes/no env value for %s", envKey)
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
