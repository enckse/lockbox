// Package inputs handles user inputs/UI elements.
package inputs

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

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
	formatTOTPEnv  = fieldTOTPEnv + "_FORMAT"
	// KeyModeEnv indicates what the KEY value is (e.g. command, plaintext).
	KeyModeEnv = prefixKey + "KEYMODE"
	// KeyEnv is the key value used by the lockbox store.
	KeyEnv = prefixKey + "KEY"
	// PlatformEnv is the platform lb is running on.
	PlatformEnv = prefixKey + "PLATFORM"
	// StoreEnv is the location of the filesystem store that lb is operating on.
	StoreEnv = prefixKey + "STORE"
	// ClipMaxEnv is the max time a value should be stored in the clipboard.
	ClipMaxEnv = prefixKey + "CLIPMAX"
	// ColorBetweenEnv is a comma-delimited list of times to color totp outputs (e.g. 0:5,30:35 which is the default).
	ColorBetweenEnv = fieldTOTPEnv + "_BETWEEN"
	// PlainKeyMode is plaintext based key resolution.
	PlainKeyMode = "plaintext"
	// CommandKeyMode will run an external command to get the key (from stdout).
	CommandKeyMode = "command"
)

// EnvOrDefault will get the environment value OR default if env is not set.
func EnvOrDefault(envKey, defaultValue string) string {
	val := os.Getenv(envKey)
	if val == "" {
		return defaultValue
	}
	return val
}

// GetKey will get the encryption key setup for lb
func GetKey() ([]byte, error) {
	useKeyMode := os.Getenv(KeyModeEnv)
	if useKeyMode == "" {
		useKeyMode = CommandKeyMode
	}
	useKey := os.Getenv(KeyEnv)
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
	case CommandKeyMode:
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
	case PlainKeyMode:
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
	case "no":
		return false, nil
	case "yes":
		return true, nil
	}
	return false, fmt.Errorf("invalid yes/no env value for %s", env)
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
	return EnvOrDefault(fieldTOTPEnv, "totp")
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
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + otpIssuer + ":" + "lbaccount",
		RawQuery: v.Encode(),
	}
	return u.String()
}
