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
	// TotpEnv allows for overriding of the special name for totp entries.
	TotpEnv = prefixKey + "TOTP"
	// ExeEnv allows for installing lb to various locations.
	ExeEnv = prefixKey + "EXE"
	// KeyModeEnv indicates what the KEY value is (e.g. command, plaintext).
	KeyModeEnv = prefixKey + "KEYMODE"
	// KeyEnv is the key value used by the lockbox store.
	KeyEnv = prefixKey + "KEY"
	// LibExecEnv is the location of libexec files for callbacks to internal exes.
	LibExecEnv = prefixKey + "LIBEXEC"
	// HooksDirEnv is the location of hooks to run before/after operations.
	HooksDirEnv = prefixKey + "HOOKDIR"
	// PlatformEnv is the platform lb is running on.
	PlatformEnv = prefixKey + "PLATFORM"
	// StoreEnv is the location of the filesystem store that lb is operating on.
	StoreEnv = prefixKey + "STORE"
	// ClipMaxEnv is the max time a value should be stored in the clipboard.
	ClipMaxEnv = prefixKey + "CLIPMAX"
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

// IsNoColorEnabled indicates if the flag is set to disable color.
func IsNoColorEnabled() (bool, error) {
	return isYesNoEnv(false, noColorEnv)
}

// IsInteractive indicates if running as a user UI experience.
func IsInteractive() (bool, error) {
	return isYesNoEnv(true, interactiveEnv)
}
