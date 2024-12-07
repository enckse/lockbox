// Package config handles user inputs/UI elements.
package config

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/seanenck/lockbox/internal/core"
	"mvdan.cc/sh/v3/shell"
)

const (
	yes               = "true"
	no                = "false"
	detectEnvironment = "detect"
	noEnvironment     = "none"
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
	exampleColorWindow = "start" + core.ColorWindowSpan + "end"
)

var (
	exampleColorWindows = []string{fmt.Sprintf("[%s]", strings.Join([]string{exampleColorWindow, exampleColorWindow, exampleColorWindow + "..."}, core.ColorWindowDelimiter))}
	configDirOffsetFile = filepath.Join(configDirName, tomlFile)
	xdgPaths            = []string{configDirOffsetFile, tomlFile}
	homePaths           = []string{filepath.Join(configDir, configDirOffsetFile), filepath.Join(configDir, tomlFile)}
	registry            = map[string]printer{}
	// TOTPDefaultColorWindow is the default coloring rules for totp
	TOTPDefaultColorWindow = []core.ColorWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = func() string {
		var results []string
		for _, w := range TOTPDefaultColorWindow {
			results = append(results, fmt.Sprintf("%d%s%d", w.Start, core.ColorWindowSpan, w.End))
		}
		return strings.Join(results, core.ColorWindowDelimiter)
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

func shlex(in string) ([]string, error) {
	return shell.Fields(in, os.Getenv)
}

func getExpand(key string) string {
	return os.ExpandEnv(os.Getenv(key))
}

func environOrDefault(envKey, defaultValue string) string {
	val := getExpand(envKey)
	if strings.TrimSpace(val) == "" {
		return defaultValue
	}
	return val
}

// NewConfigFiles will get the list of candidate config files
func NewConfigFiles() []string {
	v := EnvConfig.Get()
	if v == "" || v == noEnvironment {
		return []string{}
	}
	if err := EnvConfig.Set(noEnvironment); err != nil {
		return nil
	}
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

// IsUnset will indicate if a variable is an unset (and unset it) or return that it isn't
func IsUnset(k, v string) (bool, error) {
	if strings.TrimSpace(v) == "" {
		return true, os.Unsetenv(k)
	}
	return false, nil
}

// Environ will list the current environment keys
func Environ(set ...string) []string {
	var results []string
	filtered := len(set) > 0
	for _, k := range os.Environ() {
		for _, r := range registry {
			rawKey := r.self().Key()
			if rawKey == EnvConfig.Key() {
				continue
			}
			key := fmt.Sprintf("%s=", rawKey)
			if !strings.HasPrefix(k, key) {
				continue
			}
			if filtered {
				if !slices.Contains(set, rawKey) {
					continue
				}
			}
			results = append(results, k)
		}
	}
	sort.Strings(results)
	return results
}

// Wrap performs simple block text word wrapping
func Wrap(indent uint, in string) string {
	var sections []string
	var cur []string
	for _, line := range strings.Split(strings.TrimSpace(in), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(cur) > 0 {
				sections = append(sections, strings.Join(cur, " "))
				cur = []string{}
			}
			continue
		}
		cur = append(cur, line)
	}
	if len(cur) > 0 {
		sections = append(sections, strings.Join(cur, " "))
	}
	var out bytes.Buffer
	indenting := ""
	var cnt uint
	for cnt < indent {
		indenting = fmt.Sprintf("%s ", indenting)
		cnt++
	}
	indenture := int(80 - indent)
	for _, s := range sections {
		for _, line := range strings.Split(wrap(s, indenture), "\n") {
			fmt.Fprintf(&out, "%s%s\n", indenting, line)
		}
		fmt.Fprint(&out, "\n")
	}
	return out.String()
}

func wrap(in string, maxLength int) string {
	var lines []string
	var cur []string
	for _, p := range strings.Split(in, " ") {
		state := strings.Join(cur, " ")
		l := len(p)
		if len(state)+l >= maxLength {
			lines = append(lines, strings.Join(cur, " "))
			cur = []string{p}
		} else {
			cur = append(cur, p)
		}
	}
	if len(cur) > 0 {
		lines = append(lines, strings.Join(cur, " "))
	}
	return strings.Join(lines, "\n")
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
	override := environOrDefault(key, "")
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

// CanColor indicates if colorized output is allowed (or disabled)
func CanColor() (bool, error) {
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		return false, nil
	}
	interactive, err := EnvInteractive.Get()
	if err != nil {
		return false, err
	}
	colors := interactive
	if colors {
		isColored, err := EnvColorEnabled.Get()
		if err != nil {
			return false, err
		}
		colors = isColored
	}
	return colors, nil
}
