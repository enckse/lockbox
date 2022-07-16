package inputs

import (
	"fmt"
	"os"
	"strings"
)

func isYesNoEnv(defaultValue bool, env string) (bool, error) {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(env)))
	if len(value) == 0 {
		return defaultValue, nil
	}
	switch value {
	case "no":
		return false, nil
	case "yes":
		return true, nil
	}
	return false, fmt.Errorf("invalid yes/no env value for %s", env)
}

// IsNoClipEnabled indicates if clipboard mode is enabled.
func IsNoClipEnabled() (bool, error) {
	return isYesNoEnv(false, "LOCKBOX_NOCLIP")
}

// IsNoColorEnabled indicates if the flag is set to disable color.
func IsNoColorEnabled() (bool, error) {
	return isYesNoEnv(false, "LOCKBOX_NOCOLOR")
}

// IsInteractive indicates if running as a user UI experience.
func IsInteractive() (bool, error) {
	return isYesNoEnv(true, "LOCKBOX_INTERACTIVE")
}
