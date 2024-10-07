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
	// ModTimeFormat is the expected modtime format
	ModTimeFormat = time.RFC3339
)

var (
	// Platforms are the known platforms for lockbox
	Platforms = PlatformTypes{
		MacOSPlatform:        "macos",
		LinuxWaylandPlatform: "linux-wayland",
		LinuxXPlatform:       "linux-x",
		WindowsLinuxPlatform: "wsl",
	}
	// ReKeyFlags are the CLI argument flags for rekey handling
	ReKeyFlags = struct {
		KeyFile string
		NoKey   string
	}{"keyfile", "nokey"}
	// JSONOutputs are the JSON data output types for exporting/output of values
	JSONOutputs = JSONOutputTypes{
		Hash:  "hash",
		Blank: "empty",
		Raw:   "plaintext",
	}
	// TOTPDefaultColorWindow is the default coloring rules for totp
	TOTPDefaultColorWindow = []ColorWindow{{Start: 0, End: 5}, {Start: 30, End: 35}}
	// TOTPDefaultBetween is the default color window as a string
	TOTPDefaultBetween = func() string {
		var results []string
		for _, w := range TOTPDefaultColorWindow {
			results = append(results, fmt.Sprintf("%d%s%d", w.Start, colorWindowSpan, w.End))
		}
		return strings.Join(results, colorWindowDelimiter)
	}()
	// EnvClipMax gets the maximum clipboard time
	EnvClipMax = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(45,
				environmentBase{
					subKey: "MAX",
					cat:    clipCategory,
					desc:   "Override the amount of time before totp clears the clipboard (seconds).",
				}),
			shortDesc: "clipboard max time",
			allowZero: false,
		})
	// EnvHashLength handles the hashing output length
	EnvHashLength = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(0,
				environmentBase{
					subKey: EnvJSONDataOutput.subKey + "_HASH_LENGTH",
					desc:   fmt.Sprintf("Maximum hash string length the JSON output should contain when '%s' mode is set for JSON output.", JSONOutputs.Hash),
				}),
			shortDesc: "hash length",
			allowZero: true,
		})
	// EnvClipOSC52 indicates if OSC52 clipboard mode is enabled
	EnvClipOSC52 = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "OSC52",
					cat:    clipCategory,
					desc:   "Enable OSC52 clipboard mode.",
				}),
		})
	// EnvNoTOTP indicates if TOTP is disabled
	EnvNoTOTP = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "NOTOTP",
					desc:   "Disable TOTP integrations.",
				}),
		})
	// EnvReadOnly indicates if in read-only mode
	EnvReadOnly = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "READONLY",
					desc:   "Operate in readonly mode.",
				}),
		})
	// EnvNoClip indicates clipboard functionality is off
	EnvNoClip = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "NOCLIP",
					desc:   "Disable clipboard operations.",
				}),
		})
	// EnvDefaultCompletion disable completion detection
	EnvDefaultCompletion = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "DEFAULT_COMPLETION",
					desc:   "Use the default completion set (disable detection).",
				}),
		})
	// EnvDefaultCompletionKey is the key for default completion handling
	EnvDefaultCompletionKey = EnvDefaultCompletion.Key()
	// EnvNoColor indicates if color outputs are disabled
	EnvNoColor = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "NOCOLOR",
					desc:   "Disable terminal colors.",
				}),
		})
	// EnvNoHooks disables hooks
	EnvNoHooks = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "NOHOOKS",
					desc:   "Disable hooks",
				}),
		})
	// EnvInteractive indicates if operating in interactive mode
	EnvInteractive = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(true,
				environmentBase{
					subKey: "INTERACTIVE",
					desc:   "Enable interactive mode.",
				}),
		})
	// EnvMaxTOTP is the max TOTP time to run (default)
	EnvMaxTOTP = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(120,
				environmentBase{
					subKey: "MAX",
					cat:    totpCategory,
					desc:   "Time, in seconds, in which to show a TOTP token before automatically exiting.",
				}),
			shortDesc: "max totp time",
			allowZero: false,
		})
	// EnvTOTPToken is the leaf token to use to store TOTP tokens
	EnvTOTPToken = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("totp",
				environmentBase{
					subKey: "TOTP",
					desc:   "Attribute name to store TOTP tokens within the database.",
				}),
			allowed:    []string{"<string>"},
			canDefault: true,
		})
	// EnvPlatform is the platform that the application is running on
	EnvPlatform = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(detectedValue,
				environmentBase{
					subKey: "PLATFORM",
					desc:   "Override the detected platform.",
				}),
			allowed:    Platforms.List(),
			canDefault: false,
		})
	// EnvStore is the location of the keepass file/store
	EnvStore = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					subKey:      "STORE",
					desc:        "Directory to the database file.",
					requirement: "must be set",
				}),
			canDefault: false,
			allowed:    []string{fileExample},
		})
	// EnvHookDir is the directory of hooks to execute
	EnvHookDir = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					subKey: "HOOKDIR",
					desc:   "The path to hooks to execute on actions against the database.",
				}),
			allowed:    []string{"<directory>"},
			canDefault: true,
		})
	// EnvClipCopy allows overriding the clipboard copy command
	EnvClipCopy = environmentRegister(EnvironmentCommand{environmentBase: environmentBase{
		subKey: "COPY",
		cat:    clipCategory,
		desc:   "Override the detected platform copy command.",
	}})
	// EnvClipPaste allows overriding the clipboard paste command
	EnvClipPaste = environmentRegister(EnvironmentCommand{environmentBase: environmentBase{
		subKey: "PASTE",
		cat:    clipCategory,
		desc:   "Override the detected platform paste command.",
	}})
	// EnvTOTPColorBetween handles terminal coloring for TOTP windows (seconds)
	EnvTOTPColorBetween = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(TOTPDefaultBetween,
				environmentBase{
					subKey: "BETWEEN",
					cat:    totpCategory,
					desc: fmt.Sprintf(`Override when to set totp generated outputs to different colors,
must be a list of one (or more) rules where a '%s' delimits the start and end second (0-60 for each),
and '%s' allows for multiple windows.`, colorWindowSpan, colorWindowDelimiter),
				}),
			canDefault: true,
			allowed:    exampleColorWindows,
		})
	// EnvKeyFile is an keyfile for the database
	EnvKeyFile = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					subKey:      "KEYFILE",
					requirement: requiredKeyOrKeyFile,
					desc:        "A keyfile to access/protect the database.",
				}),
			allowed:    []string{"keyfile"},
			canDefault: true,
		})
	// EnvModTime is modtime override ability for entries
	EnvModTime = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					subKey: "SET_MODTIME",
					desc:   fmt.Sprintf("Input modification time to set for the entry\n\nExpected format: %s.", ModTimeFormat),
				}),
			canDefault: true,
			allowed:    []string{"modtime"},
		})
	// EnvJSONDataOutput controls how JSON is output in the 'data' field
	EnvJSONDataOutput = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(string(JSONOutputs.Hash),
				environmentBase{
					subKey: "JSON_DATA",
					desc:   fmt.Sprintf("Changes what the data field in JSON outputs will contain.\n\nUse '%s' with CAUTION.", JSONOutputs.Raw),
				}),
			canDefault: true,
			allowed:    JSONOutputs.List(),
		})
	// EnvFormatTOTP supports formatting the TOTP tokens for generation of tokens
	EnvFormatTOTP = environmentRegister(EnvironmentFormatter{environmentBase: environmentBase{
		subKey: "FORMAT",
		cat:    totpCategory,
		desc:   "Override the otpauth url used to store totp tokens. It must have ONE format string ('%s') to insert the totp base code.",
	}, fxn: formatterTOTP, allowed: "otpauth//url/%s/args..."})
	// EnvConfig is the location of the config file to read environment variables from
	EnvConfig = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(detectEnvironment,
				environmentBase{
					subKey: "ENV",
					desc: fmt.Sprintf(`Allows setting a specific file of environment variables for lockbox to read and use as
configuration values (an '.env' file). The keyword '%s' will disable this functionality and the keyword '%s' will
search for a file in the following paths in the user's home directory matching the first file found.

paths: %v

Note that this setting is not output as part of the environment.`, noEnvironment, detectEnvironment, detectEnvironmentPaths),
				}),
			canDefault: true,
			allowed:    []string{detectEnvironment, fileExample, noEnvironment},
		})
	// EnvKeyMode is the variable for indicating the keymode used to get the key
	EnvKeyMode = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(string(DefaultKeyMode),
				environmentBase{
					subKey:      "KEYMODE",
					requirement: "must be set to a valid mode when using a key",
					desc: fmt.Sprintf(`How to retrieve the database store password. Set to '%s' when only using a key file.
Set to '%s' to ignore the set key value`, noKeyMode, IgnoreKeyMode),
				}),
			allowed:    []string{string(AskKeyMode), string(commandKeyMode), string(IgnoreKeyMode), string(noKeyMode), string(plainKeyMode)},
			canDefault: true,
		})
	envKey = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					requirement: requiredKeyOrKeyFile,
					subKey:      "KEY",
					desc: fmt.Sprintf("The database key ('%s' mode) or command to run ('%s' mode) to retrieve the database password.",
						plainKeyMode,
						commandKeyMode),
				}),
			allowed:    []string{commandArgsExample, "password"},
			canDefault: false,
		})
	envConfigExpands = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(20,
				environmentBase{
					subKey: EnvConfig.subKey + "_EXPANDS",
					desc: `The maximum number of times to expand the input env to resolve variables (set to 0 to disable expansion).

This value can NOT be an expansion itself.`,
				}),
			shortDesc: "max expands",
			allowZero: true,
		})
	// EnvPasswordGenCount is the number of words that will be selected for password generation
	EnvPasswordGenCount = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(8,
				environmentBase{
					subKey: "COUNT",
					cat:    genCategory,
					desc:   "Number of words to select and include in the generated password.",
				}),
			shortDesc: "word count",
			allowZero: false,
		})
	// EnvPasswordGenTitle indicates if titling (e.g. uppercasing) will occur to words
	EnvPasswordGenTitle = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(true,
				environmentBase{
					subKey: "TITLE",
					cat:    genCategory,
					desc:   "Title words during password generation.",
				}),
		})
	// EnvPasswordGenTemplate is the output template for controlling how output words are placed together
	EnvPasswordGenTemplate = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(fmt.Sprintf("{{range %si, %sval := .}}{{if %si}}-{{end}}{{%sval.Text}}{{end}}", TemplateVariable, TemplateVariable, TemplateVariable, TemplateVariable),
				environmentBase{
					subKey: "TEMPLATE",
					cat:    genCategory,
					desc:   fmt.Sprintf("The go text template to use to format the chosen words into a password (use '%s' to include a '$' to avoid shell expansion issues). Fields available are Text, Position.Start, and Position.End.", TemplateVariable),
				}),
			allowed:    []string{"<go template>"},
			canDefault: true,
		})
	// EnvPasswordGenWordList is the command text to generate the word list
	EnvPasswordGenWordList = environmentRegister(EnvironmentCommand{environmentBase: environmentBase{
		subKey: "WORDLIST",
		cat:    genCategory,
		desc:   "Command to retrieve the word list to use for password generation (must be split by newline).",
	}})
	// EnvLanguage is the language to use for everything
	EnvLanguage = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("en-US",
				environmentBase{
					subKey: "LANGUAGE",
					desc:   "Language to run under.",
				}),
			allowed:    []string{"<language code>"},
			canDefault: true,
		})
	// EnvNoPasswordGen disables password generation.
	EnvNoPasswordGen = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					subKey: "NOPWGEN",
					desc:   "Disable password generation.",
				}),
		})
	// EnvPasswordGenChars allows for restricting which characters can be used
	EnvPasswordGenChars = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					subKey: "CHARS",
					cat:    genCategory,
					desc:   "The set of allowed characters in output words (empty means any characters are allowed).",
				}),
			allowed:    []string{"<list of characters>"},
			canDefault: true,
		})
)

// GetReKey will get the rekey environment settings
func GetReKey(args []string) (ReKeyArgs, error) {
	set := flag.NewFlagSet("rekey", flag.ExitOnError)
	keyFile := set.String(ReKeyFlags.KeyFile, "", "new keyfile")
	noKey := set.Bool(ReKeyFlags.NoKey, false, "disable password/key credential")
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
	for _, item := range registeredEnv {
		env := item.self()
		value, allow := item.values()
		if len(value) == 0 {
			value = "(unset)"
		}
		description := Wrap(2, env.desc)
		requirement := "optional/default"
		r := strings.TrimSpace(env.requirement)
		if r != "" {
			requirement = r
		}
		text := fmt.Sprintf("\n%s\n%s  requirement: %s\n  default: %s\n  options: %s\n", env.Key(), description, requirement, value, strings.Join(allow, "|"))
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
	val := JSONOutputMode(strings.ToLower(strings.TrimSpace(EnvJSONDataOutput.Get())))
	switch val {
	case JSONOutputs.Hash, JSONOutputs.Blank, JSONOutputs.Raw:
		return val, nil
	}
	return JSONOutputs.Blank, fmt.Errorf("invalid JSON output mode: %s", val)
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
