// Package inputs handles user inputs/UI elements.
package inputs

import (
	"fmt"
	"os"
	"strings"
)

const (
	prefixKey      = "LOCKBOX_"
	noClipEnv      = prefixKey + "NOCLIP"
	noColorEnv     = prefixKey + "NOCOLOR"
	interactiveEnv = prefixKey + "INTERACTIVE"
	gitEnabledEnv  = prefixKey + "GIT"
	gitQuietEnv    = gitEnabledEnv + "_QUIET"
	// TotpEnv allows for overriding of the special name for totp entries.
	TotpEnv = prefixKey + "TOTP"
	// KeyModeEnv indicates what the KEY value is (e.g. command, plaintext).
	KeyModeEnv = prefixKey + "KEYMODE"
	// KeyEnv is the key value used by the lockbox store.
	KeyEnv = prefixKey + "KEY"
	// HooksDirEnv is the location of hooks to run before/after operations.
	HooksDirEnv = prefixKey + "HOOKDIR"
	// PlatformEnv is the platform lb is running on.
	PlatformEnv = prefixKey + "PLATFORM"
	// StoreEnv is the location of the filesystem store that lb is operating on.
	StoreEnv = prefixKey + "STORE"
	// ClipMaxEnv is the max time a value should be stored in the clipboard.
	ClipMaxEnv = prefixKey + "CLIPMAX"
	// ColorBetweenEnv is a comma-delimited list of times to color totp outputs (e.g. 0:5,30:35 which is the default).
	ColorBetweenEnv = prefixKey + "TOTPBETWEEN"
)

// EnvOrDefault will get the environment value OR default if env is not set.
func EnvOrDefault(envKey, defaultValue string) string {
	val := os.Getenv(envKey)
	if val == "" {
		return defaultValue
	}
	return val
}

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
	return isYesNoEnv(false, noClipEnv)
}

// IsGitQuiet indicates if git operations should be 'quiet' (no stdout/stderr)
func IsGitQuiet() (bool, error) {
	return isYesNoEnv(true, gitQuietEnv)
}

// IsGitEnabled indicates if the filesystem store is a git repo
func IsGitEnabled() (bool, error) {
	return isYesNoEnv(true, gitEnabledEnv)
}

// IsNoColorEnabled indicates if the flag is set to disable color.
func IsNoColorEnabled() (bool, error) {
	return isYesNoEnv(false, noColorEnv)
}

// IsInteractive indicates if running as a user UI experience.
func IsInteractive() (bool, error) {
	return isYesNoEnv(true, interactiveEnv)
}
