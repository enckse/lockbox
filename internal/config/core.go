// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/seanenck/lockbox/internal/config/store"
	"github.com/seanenck/lockbox/internal/util"
)

const (
	yes               = "true"
	no                = "false"
	detectEnvironment = "detect"
	tomlFile          = "lockbox.toml"
	// sub categories
	clipCategory    keyCategory = "CLIP_"
	totpCategory    keyCategory = "TOTP_"
	genCategory     keyCategory = "PWGEN_"
	jsonCategory    keyCategory = "JSON_"
	credsCategory   keyCategory = "CREDENTIALS_"
	defaultCategory keyCategory = "DEFAULTS_"
	hookCategory    keyCategory = "HOOKS_"
	// YesValue are yes (on) values
	YesValue = yes
	// NoValue are no (off) values
	NoValue = no
	// TemplateVariable is used to handle '$' in shell vars (due to expansion)
	TemplateVariable     = "[%]"
	configDirName        = "lockbox"
	configDir            = ".config"
	environmentPrefix    = "LOCKBOX_"
	commandArgsExample   = "[cmd args...]"
	fileExample          = "<file>"
	detectedValue        = "<detected>"
	requiredKeyOrKeyFile = "a key, a key file, or both must be set"
	// ModTimeFormat is the expected modtime format
	ModTimeFormat      = time.RFC3339
	exampleColorWindow = "start" + util.TimeWindowSpan + "end"
)

var (
	exampleColorWindows = []string{fmt.Sprintf("[%s]", strings.Join([]string{exampleColorWindow, exampleColorWindow, exampleColorWindow + "..."}, util.TimeWindowDelimiter))}
	configDirOffsetFile = filepath.Join(configDirName, tomlFile)
	xdgPaths            = []string{configDirOffsetFile, tomlFile}
	homePaths           = []string{filepath.Join(configDir, configDirOffsetFile), filepath.Join(configDir, tomlFile)}
	registry            = map[string]printer{}
	// TOTPDefaultColorWindow is the default coloring rules for totp
	TOTPDefaultColorWindow = []util.TimeWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = func() string {
		var results []string
		for _, w := range TOTPDefaultColorWindow {
			results = append(results, fmt.Sprintf("%d%s%d", w.Start, util.TimeWindowSpan, w.End))
		}
		return strings.Join(results, util.TimeWindowDelimiter)
	}()
)

type (
	keyCategory string
	printer     interface {
		values() (string, []string)
		self() environmentBase
		toml() (tomlType, string)
	}
)

// NewConfigFiles will get the list of candidate config files
func NewConfigFiles() []string {
	v := os.Expand(os.Getenv(EnvConfig.Key()), os.Getenv)
	if v != detectEnvironment {
		return []string{v}
	}
	var options []string
	pathAdder := func(root string, err error, subs []string) {
		if err == nil && root != "" {
			for _, s := range subs {
				options = append(options, filepath.Join(root, s))
			}
		}
	}
	pathAdder(os.Getenv("XDG_CONFIG_HOME"), nil, xdgPaths)
	h, err := os.UserHomeDir()
	pathAdder(h, err, homePaths)
	return options
}

func environmentRegister[T printer](obj T) T {
	registry[obj.self().Key()] = obj
	return obj
}

func newDefaultedEnvironment[T any](val T, base environmentBase) environmentDefault[T] {
	obj := environmentDefault[T]{}
	obj.environmentBase = base
	obj.defaultValue = val
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
