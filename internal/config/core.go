// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/seanenck/lockbox/internal/config/store"
	"github.com/seanenck/lockbox/internal/util"
)

const (
	// sub categories
	clipCategory         = "CLIP_"
	totpCategory         = "TOTP_"
	genCategory          = "PWGEN_"
	jsonCategory         = "JSON_"
	credsCategory        = "CREDENTIALS_"
	defaultCategory      = "DEFAULTS_"
	hookCategory         = "HOOKS_"
	environmentPrefix    = "LOCKBOX_"
	commandArgsExample   = "[cmd args...]"
	fileExample          = "<file>"
	requiredKeyOrKeyFile = "a key, a key file, or both must be set"
	// ModTimeFormat is the expected modtime format
	ModTimeFormat      = time.RFC3339
	exampleColorWindow = "start" + util.TimeWindowSpan + "end"
	detectedValue      = "(detected)"
	unset              = "(unset)"
	arrayDelimiter     = " "
)

const (
	canDefaultFlag = iota
	canExpandFlag
	isCommandFlag
)

var (
	exampleColorWindows = []string{fmt.Sprintf("[%s]", strings.Join([]string{exampleColorWindow, exampleColorWindow, exampleColorWindow + "..."}, arrayDelimiter))}
	configDirFile       = filepath.Join("lockbox", "config.toml")
	registry            = map[string]printer{}
	// ConfigXDG is the offset to the config for XDG
	ConfigXDG = configDirFile
	// ConfigHome is the offset to the config HOME
	ConfigHome = filepath.Join(".config", configDirFile)
	// ConfigEnv allows overriding the config detection
	ConfigEnv = environmentPrefix + "CONFIG_TOML"
	// YesValue is the string variant of 'Yes' (or true) items
	YesValue = strconv.FormatBool(true)
	// NoValue is the string variant of 'No' (or false) items
	NoValue = strconv.FormatBool(false)
	// TOTPDefaultColorWindow is the default coloring rules for totp
	TOTPDefaultColorWindow = []util.TimeWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = func() []string {
		var results []string
		for _, w := range TOTPDefaultColorWindow {
			results = append(results, fmt.Sprintf("%d%s%d", w.Start, util.TimeWindowSpan, w.End))
		}
		return results
	}()
)

type (
	stringsFlags int
	printer      interface {
		display() metaData
		self() environmentBase
	}
)

// NewConfigFiles will get the list of candidate config files
func NewConfigFiles() []string {
	v := os.Expand(os.Getenv(ConfigEnv), os.Getenv)
	if v != "" {
		return []string{v}
	}
	var options []string
	pathAdder := func(root, sub string, err error) {
		if err == nil && root != "" {
			options = append(options, filepath.Join(root, sub))
		}
	}
	pathAdder(os.Getenv("XDG_CONFIG_HOME"), ConfigXDG, nil)
	h, err := os.UserHomeDir()
	pathAdder(h, ConfigHome, err)
	return options
}

func environmentRegister[T printer](obj T) T {
	registry[obj.self().Key()] = obj
	return obj
}

func newDefaultedEnvironment[T any](val T, base environmentBase) environmentDefault[T] {
	obj := environmentDefault[T]{}
	obj.environmentBase = base
	obj.value = val
	return obj
}

func formatterTOTP(key, value string) string {
	const (
		otpAuth   = "otpauth"
		otpIssuer = "lbissuer"
	)
	if strings.HasPrefix(value, otpAuth) {
		return value
	}
	override, ok := store.GetString(key)
	if ok {
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

// CanColor indicates if colorized output is allowed (or disabled)
func CanColor() (bool, error) {
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		return false, nil
	}
	colors := EnvInteractive.Get()
	if colors {
		colors = EnvColorEnabled.Get()
	}
	return colors, nil
}
