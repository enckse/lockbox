// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/config/store"
)

const (
	// sub categories
	featureCategory      = "FEATURE_"
	clipCategory         = "CLIP_"
	totpCategory         = "TOTP_"
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
	unset              = "(unset)"
	arrayDelimiter     = " "
	// TimeWindowSpan indicates the delineation between start -> end (start:end)
	TimeWindowSpan = ":"
	// NoColorFlag is the common color disable flag
	NoColorFlag = "NO_COLOR"
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
	// Loader indicates how config files should be read
	Loader interface {
		Read(string) (io.Reader, error)
		Check(string) bool
		Home() (string, error)
	}
)

// Parse will parse the config based on the loader
func Parse(loader Loader) error {
	process := func(path string) (bool, error) {
		if loader.Check(path) {
			r, err := loader.Read(path)
			if err != nil {
				return false, err
			}
			return true, Load(r, loader)
		}
		return false, nil
	}
	v := os.Expand(os.Getenv(ConfigEnv), os.Getenv)
	if v != "" {
		// NOTE: when this environment variable is set - either load the config or exit, the user does NOT want the default config options regardless
		_, err := process(v)
		return err
	}
	pathAdder := func(root, sub string) (bool, error) {
		if root != "" {
			return process(filepath.Join(root, sub))
		}
		return false, nil
	}
	ok, err := pathAdder(os.Getenv("XDG_CONFIG_HOME"), ConfigXDG)
	if ok || err != nil {
		return err
	}
	h, err := loader.Home()
	if err != nil {
		return err
	}
	_, err = pathAdder(h, ConfigHome)
	return err
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
func CanColor() bool {
	if !EnvFeatureColor.Get() {
		return false
	}
	if _, noColor := os.LookupEnv(NoColorFlag); noColor {
		return false
	}
	return true
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
	return strings.Join(readNested(reflect.TypeFor[Word](), ""), ", ")
}

// NewFeatureError creates an error if a feature is not enabled
func NewFeatureError(name string) error {
	return fmt.Errorf("%s feature is disabled", name)
}
