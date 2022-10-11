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
	clipBaseEnv    = prefixKey + "CLIP_"
	formatTOTPEnv  = fieldTOTPEnv + "_FORMAT"
	keyModeEnv     = prefixKey + "KEYMODE"
	keyEnv         = prefixKey + "KEY"
	plainKeyMode   = "plaintext"
	commandKeyMode = "command"
	// PlatformEnv is the platform lb is running on.
	PlatformEnv = prefixKey + "PLATFORM"
	// StoreEnv is the location of the filesystem store that lb is operating on.
	StoreEnv = prefixKey + "STORE"
	// ClipMaxEnv is the max time a value should be stored in the clipboard.
	ClipMaxEnv = clipBaseEnv + "MAX"
	// ColorBetweenEnv is a comma-delimited list of times to color totp outputs (e.g. 0:5,30:35 which is the default).
	ColorBetweenEnv = fieldTOTPEnv + "_BETWEEN"
	// ClipPasteEnv allows overriding the clipboard paste command
	ClipPasteEnv = clipBaseEnv + "PASTE"
	// ClipCopyEnv allows overriding the clipboard copy command
	ClipCopyEnv = clipBaseEnv + "COPY"
	// DefaultsCommand will get the environment values WITHOUT current environment settings
	DefaultsCommand  = "-defaults"
	isYes            = "yes"
	isNo             = "no"
	defaultTOTPField = "totp"
)

var (
	isYesNoArgs = []string{isYes, isNo}
)

type (
	environmentOutput struct {
		showValues bool
	}
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

func (o environmentOutput) printEnvironmentVariable(required bool, name, val, desc string, allowed []string) {
	value := val
	if o.showValues {
		value = os.Getenv(name)
	}
	if len(value) == 0 {
		value = "(unset)"
	}
	fmt.Printf("\n%s\n  %s\n\n  required: %t\n  value: %s\n", name, desc, required, value)
}

// ListEnvironmentVariables will print information about env variables and potential/set values
func ListEnvironmentVariables(args []string) error {
	showValues := true
	switch len(args) {
	case 0:
		break
	case 1:
		if args[0] == DefaultsCommand {
			showValues = false
		} else {
			return errors.New("unknown argument")
		}
	default:
		return errors.New("too many arguments")
	}
	e := environmentOutput{showValues: showValues}
	e.printEnvironmentVariable(true, StoreEnv, "", "directory to the database file", nil)
	e.printEnvironmentVariable(true, keyModeEnv, commandKeyMode, "how to retrieve the database store password", []string{commandKeyMode, plainKeyMode})
	e.printEnvironmentVariable(true, keyEnv, "unset", fmt.Sprintf("the database key (%s) or shell command to run (%s) to retrieve the database password", plainKeyMode, commandKeyMode), nil)
	e.printEnvironmentVariable(false, noClipEnv, isNo, "disable clipboard operations", isYesNoArgs)
	e.printEnvironmentVariable(false, noColorEnv, isNo, "disable terminal colors", isYesNoArgs)
	e.printEnvironmentVariable(false, interactiveEnv, isYes, "enable interactive mode", isYesNoArgs)
	e.printEnvironmentVariable(false, readOnlyEnv, isNo, "operate in readonly mode", isYesNoArgs)
	e.printEnvironmentVariable(false, fieldTOTPEnv, defaultTOTPField, "attribute name to store TOTP tokens within the database", nil)
	e.printEnvironmentVariable(false, formatTOTPEnv, "", "override the otpauth url used to store totp tokens (e.g. otpauth://totp/%s/rest/of/string), must have ONE format '%s' to insert the totp base code", nil)
	e.printEnvironmentVariable(false, ColorBetweenEnv, "", "override when to set totp generated outputs to different colors (e.g. 0:5,30:35), must be a list of one (or more) rules where a semicolon delimits the start and end second (0-60 for each)", nil)
	e.printEnvironmentVariable(false, ClipPasteEnv, "", "override the detected platform paste command", nil)
	e.printEnvironmentVariable(false, ClipPasteEnv, "", "override the detected platform copy command", nil)
	e.printEnvironmentVariable(false, ClipMaxEnv, "", "override the amount of time before totp clears the clipboard (e.g. 10), must be an integer", nil)
	e.printEnvironmentVariable(false, PlatformEnv, "", "override the detected platform", nil)
	return nil
}
