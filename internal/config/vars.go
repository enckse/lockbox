// Package config handles user inputs/UI elements.
package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	commandArgsExample   = "[cmd args...]"
	fileExample          = "<file>"
	detectedValue        = "<detected>"
	requiredKeyOrKeyFile = "a key, a key file, or both must be set"
	askProfile           = "ask"
	roProfile            = "readonly"
	noTOTPProfile        = "nototp"
	noClipProfile        = "noclip"
	// ModTimeFormat is the expected modtime format
	ModTimeFormat = time.RFC3339
	// JSONDataOutputHash means output data is hashed
	JSONDataOutputHash JSONOutputMode = "hash"
	// JSONDataOutputBlank means an empty entry is set
	JSONDataOutputBlank JSONOutputMode = "empty"
	// JSONDataOutputRaw means the RAW (unencrypted) value is displayed
	JSONDataOutputRaw JSONOutputMode = "plaintext"
)

var (
	registry = []printer{EnvStore, envKeyMode, envKey, EnvNoClip, EnvNoColor, EnvInteractive, EnvReadOnly, EnvTOTPToken, EnvFormatTOTP, EnvMaxTOTP, EnvTOTPColorBetween, EnvClipPaste, EnvClipCopy, EnvClipMax, EnvPlatform, EnvNoTOTP, EnvHookDir, EnvClipOSC52, EnvKeyFile, EnvModTime, EnvJSONDataOutput, EnvHashLength, EnvConfig, envConfigExpands, EnvDefaultCompletion, EnvNoHooks}
	// Platforms represent the platforms that lockbox understands to run on
	Platforms = []string{MacOSPlatform, WindowsLinuxPlatform, LinuxXPlatform, LinuxWaylandPlatform}
	// TOTPDefaultColorWindow is the default coloring rules for totp
	TOTPDefaultColorWindow = []ColorWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = toString(TOTPDefaultColorWindow)
	// EnvClipMax gets the maximum clipboard time
	EnvClipMax = EnvironmentInt{
		environmentBase: environmentBase{
			subKey: "MAX",
			cat:    clipCategory,
			desc:   "Override the amount of time before totp clears the clipboard (seconds).",
		},
		shortDesc: "clipboard max time", allowZero: false, defaultValue: 45,
	}
	// EnvHashLength handles the hashing output length
	EnvHashLength = EnvironmentInt{
		environmentBase: environmentBase{
			subKey: EnvJSONDataOutput.subKey + "_HASH_LENGTH",
			desc:   fmt.Sprintf("Maximum hash string length the JSON output should contain when '%s' mode is set for JSON output.", JSONDataOutputHash),
		},
		shortDesc: "hash length", allowZero: true, defaultValue: 0,
	}
	// EnvClipOSC52 indicates if OSC52 clipboard mode is enabled
	EnvClipOSC52 = EnvironmentBool{environmentBase: environmentBase{
		subKey: "OSC52",
		cat:    clipCategory,
		desc:   "Enable OSC52 clipboard mode.",
	}, defaultValue: false}
	// EnvNoTOTP indicates if TOTP is disabled
	EnvNoTOTP = EnvironmentBool{
		environmentBase: environmentBase{
			subKey: "NOTOTP",
			desc:   "Disable TOTP integrations.",
		},
		defaultValue: false,
	}
	// EnvReadOnly indicates if in read-only mode
	EnvReadOnly = EnvironmentBool{environmentBase: environmentBase{
		subKey: "READONLY",
		desc:   "Operate in readonly mode.",
	}, defaultValue: false}
	// EnvNoClip indicates clipboard functionality is off
	EnvNoClip = EnvironmentBool{
		environmentBase: environmentBase{
			subKey: "NOCLIP",
			desc:   "Disable clipboard operations.",
		},
		defaultValue: false,
	}
	// EnvDefaultCompletion disable completion detection
	EnvDefaultCompletion = EnvironmentBool{
		environmentBase: environmentBase{
			subKey: "DEFAULT_COMPLETION",
			desc:   "Use the default completion set (disable detection).",
		},
		defaultValue: false,
	}
	// EnvDefaultCompletionKey is the key for default completion handling
	EnvDefaultCompletionKey = EnvDefaultCompletion.key()
	// EnvNoColor indicates if color outputs are disabled
	EnvNoColor = EnvironmentBool{environmentBase: environmentBase{
		subKey: "NOCOLOR",
		desc:   "Disable terminal colors.",
	}, defaultValue: false}
	// EnvNoHooks disables hooks
	EnvNoHooks = EnvironmentBool{environmentBase: environmentBase{
		subKey: "NOHOOKS",
		desc:   "Disable hooks",
	}, defaultValue: false}
	// EnvInteractive indicates if operating in interactive mode
	EnvInteractive = EnvironmentBool{environmentBase: environmentBase{
		subKey: "INTERACTIVE",
		desc:   "Enable interactive mode.",
	}, defaultValue: true}
	// EnvMaxTOTP is the max TOTP time to run (default)
	EnvMaxTOTP = EnvironmentInt{environmentBase: environmentBase{
		subKey: "MAX",
		cat:    totpCategory,
		desc:   "Time, in seconds, in which to show a TOTP token before automatically exiting.",
	}, shortDesc: "max totp time", allowZero: false, defaultValue: 120}
	// EnvTOTPToken is the leaf token to use to store TOTP tokens
	EnvTOTPToken = EnvironmentString{environmentBase: environmentBase{
		subKey: "TOTP",
		desc:   "Attribute name to store TOTP tokens within the database.",
	}, allowed: []string{"<string>"}, canDefault: true, defaultValue: "totp"}
	// EnvPlatform is the platform that the application is running on
	EnvPlatform = EnvironmentString{environmentBase: environmentBase{
		subKey: "PLATFORM",
		desc:   "Override the detected platform.",
	}, defaultValue: detectedValue, allowed: Platforms, canDefault: false}
	// EnvStore is the location of the keepass file/store
	EnvStore = EnvironmentString{environmentBase: environmentBase{
		subKey: "STORE",
		desc:   "Directory to the database file.", requirement: "must be set",
	}, canDefault: false, allowed: []string{fileExample}}
	// EnvHookDir is the directory of hooks to execute
	EnvHookDir = EnvironmentString{environmentBase: environmentBase{
		subKey: "HOOKDIR",
		desc:   "The path to hooks to execute on actions against the database.",
	}, allowed: []string{"<directory>"}, canDefault: true, defaultValue: ""}
	// EnvClipCopy allows overriding the clipboard copy command
	EnvClipCopy = EnvironmentCommand{environmentBase: environmentBase{
		subKey: "COPY",
		cat:    clipCategory,
		desc:   "Override the detected platform copy command.",
	}}
	// EnvClipPaste allows overriding the clipboard paste command
	EnvClipPaste = EnvironmentCommand{environmentBase: environmentBase{
		subKey: "PASTE",
		cat:    clipCategory,
		desc:   "Override the detected platform paste command.",
	}}
	// EnvTOTPColorBetween handles terminal coloring for TOTP windows (seconds)
	EnvTOTPColorBetween = EnvironmentString{environmentBase: environmentBase{
		subKey: "BETWEEN",
		cat:    totpCategory,
		desc: fmt.Sprintf(`Override when to set totp generated outputs to different colors,
must be a list of one (or more) rules where a '%s' delimits the start and end second (0-60 for each),
and '%s' allows for multiple windows.`, colorWindowSpan, colorWindowDelimiter),
	}, canDefault: true, defaultValue: TOTPDefaultBetween, allowed: exampleColorWindows}
	// EnvKeyFile is an keyfile for the database
	EnvKeyFile = EnvironmentString{environmentBase: environmentBase{
		subKey: "KEYFILE", requirement: requiredKeyOrKeyFile,
		desc: "A keyfile to access/protect the database.",
	}, allowed: []string{"keyfile"}, canDefault: true, defaultValue: ""}
	// EnvModTime is modtime override ability for entries
	EnvModTime = EnvironmentString{environmentBase: environmentBase{
		subKey: "SET_MODTIME",
		desc:   fmt.Sprintf("Input modification time to set for the entry\n\nExpected format: %s.", ModTimeFormat),
	}, canDefault: true, defaultValue: "", allowed: []string{"modtime"}}
	// EnvJSONDataOutput controls how JSON is output in the 'data' field
	EnvJSONDataOutput = EnvironmentString{
		environmentBase: environmentBase{
			subKey: "JSON_DATA",
			desc:   fmt.Sprintf("Changes what the data field in JSON outputs will contain.\n\nUse '%s' with CAUTION.", JSONDataOutputRaw),
		}, canDefault: true, defaultValue: string(JSONDataOutputHash),
		allowed: []string{string(JSONDataOutputRaw), string(JSONDataOutputHash), string(JSONDataOutputBlank)},
	}
	// EnvFormatTOTP supports formatting the TOTP tokens for generation of tokens
	EnvFormatTOTP = EnvironmentFormatter{environmentBase: environmentBase{
		subKey: "FORMAT",
		cat:    totpCategory,
		desc:   "Override the otpauth url used to store totp tokens. It must have ONE format string ('%s') to insert the totp base code.",
	}, fxn: formatterTOTP, allowed: "otpauth//url/%s/args..."}
	// EnvConfig is the location of the config file to read environment variables from
	EnvConfig = EnvironmentString{environmentBase: environmentBase{
		subKey: "ENV",
		desc: fmt.Sprintf(`Allows setting a specific file of environment variables for lockbox to read and use as
configuration values (an '.env' file). The keyword '%s' will disable this functionality and the keyword '%s' will
search for a file in the following paths in the user's home directory matching the first file found.

paths: %v

Note that this setting is not output as part of the environment.`, noEnvironment, detectEnvironment, detectEnvironmentPaths),
	}, canDefault: true, defaultValue: detectEnvironment, allowed: []string{detectEnvironment, fileExample, noEnvironment}}
	envKeyMode = EnvironmentString{
		environmentBase: environmentBase{
			subKey: "KEYMODE", requirement: "must be set to a valid mode when using a key",
			desc: fmt.Sprintf(`How to retrieve the database store password. Set to '%s' when only using a key file.
Set to '%s' to ignore the set key value`, noKeyMode, IgnoreKeyMode), whenUnset: string(DefaultKeyMode),
		},
		allowed:    []string{string(askKeyMode), string(commandKeyMode), string(IgnoreKeyMode), string(noKeyMode), string(plainKeyMode)},
		canDefault: true, defaultValue: "",
	}
	envKey = EnvironmentString{environmentBase: environmentBase{
		requirement: requiredKeyOrKeyFile, subKey: "KEY",
		desc: fmt.Sprintf("The database key ('%s' mode) or command to run ('%s' mode) to retrieve the database password.",
			plainKeyMode,
			commandKeyMode),
	}, allowed: []string{commandArgsExample, "password"}, canDefault: false}
	envConfigExpands = EnvironmentInt{environmentBase: environmentBase{
		subKey: EnvConfig.subKey + "_EXPANDS",
		desc: `The maximum number of times to expand the input env to resolve variables (set to 0 to disable expansion).

This value can NOT be an expansion itself.`,
	}, shortDesc: "max expands", allowZero: true, defaultValue: 20}
)

// GetReKey will get the rekey environment settings
func GetReKey(args []string) (ReKeyArgs, error) {
	set := flag.NewFlagSet("rekey", flag.ExitOnError)
	keyFile := set.String(ReKeyKeyFileFlag, "", "new keyfile")
	noKey := set.Bool(ReKeyNoKeyFlag, false, "disable password/key credential")
	if err := set.Parse(args); err != nil {
		return ReKeyArgs{}, err
	}
	noPass := *noKey
	file := *keyFile
	if strings.TrimSpace(file) == "" && noPass {
		return ReKeyArgs{}, errors.New("a key or keyfile must be passed for rekey")
	}
	return ReKeyArgs{KeyFile: file, NoKey: noPass}, nil
}

// ListEnvironmentVariables will print information about env variables
func ListEnvironmentVariables() []string {
	var results []string
	for _, item := range registry {
		env := item.self()
		value, allow := item.values()
		if len(value) == 0 {
			value = "(unset)"
			if env.whenUnset != "" {
				value = env.whenUnset
			}
		}
		description := Wrap(2, env.desc)
		requirement := "optional/default"
		r := strings.TrimSpace(env.requirement)
		if r != "" {
			requirement = r
		}
		text := fmt.Sprintf("\n%s\n%s  requirement: %s\n  default: %s\n  options: %s\n", env.key(), description, requirement, value, strings.Join(allow, "|"))
		results = append(results, text)
	}
	sort.Strings(results)
	return results
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

// ParseJSONOutput handles detecting the JSON output mode
func ParseJSONOutput() (JSONOutputMode, error) {
	val := strings.ToLower(strings.TrimSpace(EnvJSONDataOutput.Get()))
	switch JSONOutputMode(val) {
	case JSONDataOutputHash:
		return JSONDataOutputHash, nil
	case JSONDataOutputBlank:
		return JSONDataOutputBlank, nil
	case JSONDataOutputRaw:
		return JSONDataOutputRaw, nil
	}
	return JSONDataOutputBlank, fmt.Errorf("invalid JSON output mode: %s", val)
}

func exportProfileKeyValue(e environmentBase, val string) string {
	return fmt.Sprintf("\"$%s\" = \"%s\"", e.key(), val)
}

func newProfile(keys []string) CompletionProfile {
	p := CompletionProfile{}
	p.Clip = true
	p.List = true
	p.TOTP = true
	p.Write = true
	name := ""
	sort.Strings(keys)
	var e []string
	for _, k := range keys {
		name = fmt.Sprintf("%s%s-", name, k)
		switch k {
		case askProfile:
			e = append(e, exportProfileKeyValue(envKeyMode.environmentBase, string(askKeyMode)))
			p.List = false
		case noTOTPProfile:
			e = append(e, exportProfileKeyValue(EnvNoTOTP.environmentBase, yes))
			p.TOTP = false
		case noClipProfile:
			e = append(e, exportProfileKeyValue(EnvNoClip.environmentBase, yes))
			p.Clip = false
		case roProfile:
			e = append(e, exportProfileKeyValue(EnvReadOnly.environmentBase, yes))
			p.Write = false
		}
	}
	sort.Strings(e)
	p.Env = e
	p.Name = strings.TrimSuffix(name, "-")
	return p
}

func generateProfiles(keys []string) map[string]CompletionProfile {
	m := make(map[string]CompletionProfile)
	if len(keys) == 0 {
		return m
	}
	p := newProfile(keys)
	m[p.Name] = p
	for _, cur := range keys {
		var subset []string
		for _, key := range keys {
			if key == cur {
				continue
			}
			subset = append(subset, key)
		}

		for _, p := range generateProfiles(subset) {
			m[p.Name] = p
		}
	}
	return m
}

// LoadCompletionProfiles will generate known completion profile with backing env information
func LoadCompletionProfiles() []CompletionProfile {
	loaded := generateProfiles([]string{noClipProfile, roProfile, noTOTPProfile, askProfile})
	var profiles []CompletionProfile
	for _, v := range loaded {
		profiles = append(profiles, v)
	}
	sort.Slice(profiles, func(i, j int) bool {
		return strings.Compare(profiles[i].Name, profiles[j].Name) < 0
	})
	profiles = append(profiles, CompletionProfile{Clip: true, Write: true, TOTP: true, List: true, Default: true})
	return profiles
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
		isColored, err := EnvNoColor.Get()
		if err != nil {
			return false, err
		}
		colors = !isColored
	}
	return colors, nil
}
