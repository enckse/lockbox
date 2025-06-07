// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

const (
	// sub categories
	clipCategory         = "CLIP_"
	totpCategory         = "TOTP_"
	genCategory          = "PWGEN_"
	jsonCategory         = "JSON_"
	credsCategory        = "CREDENTIALS_"
	defaultCategory      = "DEFAULTS_"
	environmentPrefix    = "LOCKBOX_"
	commandArgsExample   = "[cmd args...]"
	fileExample          = "<file>"
	requiredKeyOrKeyFile = "a key, a key file, or both must be set"
	// ModTimeFormat is the expected modtime format
	ModTimeFormat      = time.RFC3339
	exampleColorWindow = "start" + TimeWindowSpan + "end"
	detectedValue      = "(detected)"
	unset              = "(unset)"
	arrayDelimiter     = " "
	// TimeWindowSpan indicates the delineation between start -> end (start:end)
	TimeWindowSpan = ":"
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
	TOTPDefaultColorWindow = []TimeWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = func() []string {
		var results []string
		for _, w := range TOTPDefaultColorWindow {
			results = append(results, fmt.Sprintf("%d%s%d", w.Start, TimeWindowSpan, w.End))
		}
		return results
	}()
)

type (
	// TimeWindow for handling terminal colors based on timing
	TimeWindow struct {
		Start int
		End   int
	}

	stringsFlags int
	printer      interface {
		display() metaData
		self() environmentBase
	}
	// Position is the start/end of a word in a greater set
	Position struct {
		Start int
		End   int
	}
	// Word is the text and position in a greater position
	Word struct {
		Text     string
		Position Position
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

func readNested(v reflect.Type, root string) []string {
	var fields []string
	for i := range v.NumField() {
		field := v.Field(i)
		if field.Type.Kind() == reflect.Struct {
			fields = append(fields, readNested(field.Type, fmt.Sprintf("%s.", field.Name))...)
		} else {
			fields = append(fields, fmt.Sprintf("%s%s", root, field.Name))
		}
	}
	return fields
}

// TextPositionFields is the displayable set of templated fields
func TextPositionFields() string {
	return strings.Join(readNested(reflect.TypeOf(Word{}), ""), ", ")
}
