// Package config handles user inputs/UI elements.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/shell"
)

const (
	colorWindowDelimiter = " "
	colorWindowSpan      = ":"
	exampleColorWindow   = "start" + colorWindowSpan + "end"
	yes                  = "yes"
	no                   = "no"
	detectEnvironment    = "detect"
	noEnvironment        = "none"
	tomlFile             = "lockbox.toml"
	unknownPlatform      = ""
	// sub categories
	clipCategory keyCategory = "CLIP_"
	totpCategory keyCategory = "TOTP_"
	genCategory  keyCategory = "PWGEN_"
	// YesValue are yes (on) values
	YesValue = yes
	// TemplateVariable is used to handle '$' in shell vars (due to expansion)
	TemplateVariable  = "[%]"
	configDirName     = "lockbox"
	configDir         = ".config"
	environmentPrefix = "LOCKBOX_"
)

var (
	configDirOffsetFile = filepath.Join(configDirName, tomlFile)
	xdgPaths            = []string{configDirOffsetFile, tomlFile}
	homePaths           = []string{filepath.Join(configDir, configDirOffsetFile), filepath.Join(configDir, tomlFile)}
	exampleColorWindows = []string{strings.Join([]string{exampleColorWindow, exampleColorWindow, exampleColorWindow + "..."}, colorWindowDelimiter)}
	registeredEnv       = []printer{}
)

type (
	keyCategory string
	// JSONOutputMode is the output mode definition
	JSONOutputMode string
	// SystemPlatform represents the platform lockbox is running on.
	SystemPlatform  string
	environmentBase struct {
		subKey      string
		cat         keyCategory
		desc        string
		requirement string
	}
	environmentDefault[T any] struct {
		environmentBase
		defaultValue T
	}
	// EnvironmentInt are environment settings that are integers
	EnvironmentInt struct {
		environmentDefault[int]
		allowZero bool
		shortDesc string
	}
	// EnvironmentBool are environment settings that are booleans
	EnvironmentBool struct {
		environmentDefault[bool]
	}
	// EnvironmentString are string-based settings
	EnvironmentString struct {
		environmentDefault[string]
		canDefault bool
		allowed    []string
	}
	// EnvironmentCommand are settings that are parsed as shell commands
	EnvironmentCommand struct {
		environmentBase
	}
	// EnvironmentFormatter allows for sending a string into a get request
	EnvironmentFormatter struct {
		environmentBase
		allowed string
		fxn     func(string, string) string
	}
	printer interface {
		values() (string, []string)
		self() environmentBase
	}
	// ColorWindow for handling terminal colors based on timing
	ColorWindow struct {
		Start int
		End   int
	}
	// ReKeyArgs are the arguments for rekeying
	ReKeyArgs struct {
		NoKey   bool
		KeyFile string
	}
	// PlatformTypes defines systems lockbox is known to run on or can run on
	PlatformTypes struct {
		MacOSPlatform        SystemPlatform
		LinuxWaylandPlatform SystemPlatform
		LinuxXPlatform       SystemPlatform
		WindowsLinuxPlatform SystemPlatform
	}
	// JSONOutputTypes indicate how JSON data can be exported for values
	JSONOutputTypes struct {
		Hash  JSONOutputMode
		Blank JSONOutputMode
		Raw   JSONOutputMode
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

func (e environmentBase) Key() string {
	return fmt.Sprintf(environmentPrefix+"%s%s", string(e.cat), e.subKey)
}

// Get will get the boolean value for the setting
func (e EnvironmentBool) Get() (bool, error) {
	return parseStringYesNo(e, getExpand(e.Key()))
}

func parseStringYesNo(e EnvironmentBool, in string) (bool, error) {
	read := strings.ToLower(strings.TrimSpace(in))
	switch read {
	case no:
		return false, nil
	case yes:
		return true, nil
	case "":
		return e.defaultValue, nil
	}

	return false, fmt.Errorf("invalid yes/no env value for %s", e.Key())
}

// Get will get the integer value for the setting
func (e EnvironmentInt) Get() (int, error) {
	val := e.defaultValue
	use := getExpand(e.Key())
	if use != "" {
		i, err := strconv.Atoi(use)
		if err != nil {
			return -1, err
		}
		invalid := false
		check := ""
		if e.allowZero {
			check = "="
		}
		switch i {
		case 0:
			invalid = !e.allowZero
		default:
			invalid = i < 0
		}
		if invalid {
			return -1, fmt.Errorf("%s must be >%s 0", e.shortDesc, check)
		}
		val = i
	}
	return val, nil
}

// Get will read the string from the environment
func (e EnvironmentString) Get() string {
	if !e.canDefault {
		return getExpand(e.Key())
	}
	return environOrDefault(e.Key(), e.defaultValue)
}

// Get will read (and shlex) the value if set
func (e EnvironmentCommand) Get() ([]string, error) {
	value := environOrDefault(e.Key(), "")
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	return shlex(value)
}

// KeyValue will get the string representation of the key+value
func (e environmentBase) KeyValue(value string) string {
	return fmt.Sprintf("%s=%s", e.Key(), value)
}

// Setenv will do an environment set for the value to key
func (e environmentBase) Set(value string) error {
	unset, err := IsUnset(e.Key(), value)
	if err != nil {
		return err
	}
	if unset {
		return nil
	}
	return os.Setenv(e.Key(), value)
}

// Get will retrieve the value with the formatted input included
func (e EnvironmentFormatter) Get(value string) string {
	return e.fxn(e.Key(), value)
}

func (e EnvironmentString) values() (string, []string) {
	return e.defaultValue, e.allowed
}

func (e environmentBase) self() environmentBase {
	return e
}

func (e EnvironmentBool) values() (string, []string) {
	val := no
	if e.defaultValue {
		val = yes
	}
	return val, []string{yes, no}
}

func (e EnvironmentInt) values() (string, []string) {
	return fmt.Sprintf("%d", e.defaultValue), []string{"<integer>"}
}

func (e EnvironmentFormatter) values() (string, []string) {
	return strings.ReplaceAll(strings.ReplaceAll(EnvFormatTOTP.Get("%s"), "%25s", "%s"), "&", " \\\n           &"), []string{e.allowed}
}

func (e EnvironmentCommand) values() (string, []string) {
	return detectedValue, []string{commandArgsExample}
}

// NewPlatform gets a new system platform.
func NewPlatform() (SystemPlatform, error) {
	env := EnvPlatform.Get()
	if env != "" {
		for _, p := range EnvPlatform.allowed {
			if p == env {
				return SystemPlatform(p), nil
			}
		}
		return unknownPlatform, errors.New("unknown platform mode")
	}
	b, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return unknownPlatform, err
	}
	raw := strings.ToLower(strings.TrimSpace(string(b)))
	parts := strings.Split(raw, " ")
	switch parts[0] {
	case "darwin":
		return Platforms.MacOSPlatform, nil
	case "linux":
		if strings.Contains(raw, "microsoft-standard-wsl") {
			return Platforms.WindowsLinuxPlatform, nil
		}
		if strings.TrimSpace(getExpand("WAYLAND_DISPLAY")) == "" {
			if strings.TrimSpace(getExpand("DISPLAY")) == "" {
				return unknownPlatform, errors.New("unable to detect linux clipboard mode")
			}
			return Platforms.LinuxXPlatform, nil
		}
		return Platforms.LinuxWaylandPlatform, nil
	}
	return unknownPlatform, errors.New("unable to detect clipboard mode")
}

// ParseColorWindow will handle parsing a window of colors for TOTP operations
func ParseColorWindow(windowString string) ([]ColorWindow, error) {
	var rules []ColorWindow
	for _, item := range strings.Split(windowString, colorWindowDelimiter) {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		parts := strings.Split(line, colorWindowSpan)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid colorization rule found: %s", line)
		}
		s, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		e, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if s < 0 || e < 0 || e < s || s > 59 || e > 59 {
			return nil, fmt.Errorf("invalid time found for colorization rule: %s", line)
		}
		rules = append(rules, ColorWindow{Start: s, End: e})
	}
	if len(rules) == 0 {
		return nil, errors.New("invalid colorization rules for totp, none found")
	}
	return rules, nil
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
func Environ() []string {
	var results []string
	for _, k := range os.Environ() {
		for _, r := range registeredEnv {
			key := r.self().Key()
			if key == EnvConfig.Key() {
				continue
			}
			key = fmt.Sprintf("%s=", key)
			if strings.HasPrefix(k, key) {
				results = append(results, k)
				break
			}
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
	registeredEnv = append(registeredEnv, obj)
	return obj
}

// List will list the platform types on the struct
func (p PlatformTypes) List() []string {
	return listFields[SystemPlatform](p)
}

// List will list the output modes on the struct
func (p JSONOutputTypes) List() []string {
	return listFields[JSONOutputMode](p)
}

func listFields[T SystemPlatform | JSONOutputMode](p any) []string {
	v := reflect.ValueOf(p)
	var vals []string
	for i := 0; i < v.NumField(); i++ {
		vals = append(vals, fmt.Sprintf("%v", v.Field(i).Interface().(T)))
	}
	sort.Strings(vals)
	return vals
}

func newDefaultedEnvironment[T any](val T, base environmentBase) environmentDefault[T] {
	obj := environmentDefault[T]{}
	obj.environmentBase = base
	obj.defaultValue = val
	return obj
}
