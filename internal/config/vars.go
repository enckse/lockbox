// Package config handles user inputs/UI elements.
package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
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
	// EnvClipTimeout gets the maximum clipboard time
	EnvClipTimeout = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(45,
				environmentBase{
					subKey: "TIMEOUT",
					cat:    clipCategory,
					desc:   "Override the amount of time before totp clears the clipboard (seconds).",
				}),
			shortDesc: "clipboard max time",
			allowZero: false,
		})
	// EnvJSONHashLength handles the hashing output length
	EnvJSONHashLength = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(0,
				environmentBase{
					cat:    jsonCategory,
					subKey: "HASH_LENGTH",
					desc:   fmt.Sprintf("Maximum string length of the JSON value when '%s' mode is set for JSON output.", JSONOutputs.Hash),
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
	// EnvTOTPEnabled indicates if TOTP is allowed
	EnvTOTPEnabled = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(true,
				environmentBase{
					cat:    totpCategory,
					subKey: "ENABLED",
					desc:   "Enable TOTP integrations.",
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
	// EnvClipEnabled indicates if clipboard is enabled
	EnvClipEnabled = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(true,
				environmentBase{
					cat:    clipCategory,
					subKey: "ENABLED",
					desc:   "Enable clipboard operations.",
				}),
		})
	// EnvDefaultCompletion disable completion detection
	EnvDefaultCompletion = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(false,
				environmentBase{
					cat:    defaultCategory,
					subKey: "COMPLETION",
					desc:   "Use the default completion set (disable detection).",
				}),
		})
	// EnvDefaultCompletionKey is the key for default completion handling
	EnvDefaultCompletionKey = EnvDefaultCompletion.Key()
	// EnvColorEnabled indicates if colors are enabled
	EnvColorEnabled = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(true,
				environmentBase{
					subKey: "COLOR_ENABLED",
					desc:   "Enable terminal colors.",
				}),
		})
	// EnvHooksEnabled indicates if hooks are enabled
	EnvHooksEnabled = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(true,
				environmentBase{
					cat:    hookCategory,
					subKey: "ENABLED",
					desc:   "Enable hooks",
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
	// EnvTOTPTimeout indicates when TOTP display should timeout
	EnvTOTPTimeout = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(120,
				environmentBase{
					subKey: "TIMEOUT",
					cat:    totpCategory,
					desc:   "Time, in seconds, to show a TOTP token before automatically exiting.",
				}),
			shortDesc: "max totp time",
			allowZero: false,
		})
	// EnvTOTPEntry is the leaf token to use to store TOTP tokens
	EnvTOTPEntry = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("totp",
				environmentBase{
					cat:    totpCategory,
					subKey: "ENTRY",
					desc:   "Entry name to store TOTP tokens within the database.",
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
					cat:    hookCategory,
					subKey: "DIRECTORY",
					desc:   "The path to hooks to execute on actions against the database.",
				}),
			allowed:    []string{"<directory>"},
			canDefault: true,
		})
	// EnvClipCopy allows overriding the clipboard copy command
	EnvClipCopy = environmentRegister(EnvironmentCommand{environmentBase: environmentBase{
		subKey: "COPY_COMMAND",
		cat:    clipCategory,
		desc:   "Override the detected platform copy command.",
	}})
	// EnvClipPaste allows overriding the clipboard paste command
	EnvClipPaste = environmentRegister(EnvironmentCommand{environmentBase: environmentBase{
		subKey: "PASTE_COMMAND",
		cat:    clipCategory,
		desc:   "Override the detected platform paste command.",
	}})
	// EnvTOTPColorBetween handles terminal coloring for TOTP windows (seconds)
	EnvTOTPColorBetween = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(TOTPDefaultBetween,
				environmentBase{
					subKey: "COLOR_WINDOWS",
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
					cat:         credsCategory,
					subKey:      "KEY_FILE",
					requirement: requiredKeyOrKeyFile,
					desc:        "A keyfile to access/protect the database.",
				}),
			allowed:    []string{"keyfile"},
			canDefault: true,
		})
	// EnvDefaultModTime is modtime override ability for entries
	EnvDefaultModTime = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					cat:    defaultCategory,
					subKey: "MODTIME",
					desc:   fmt.Sprintf("Input modification time to set for the entry\n\nExpected format: %s.", ModTimeFormat),
				}),
			canDefault: true,
			allowed:    []string{"modtime"},
		})
	// EnvJSONMode controls how JSON is output in the 'data' field
	EnvJSONMode = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(string(JSONOutputs.Hash),
				environmentBase{
					cat:    jsonCategory,
					subKey: "MODE",
					desc:   fmt.Sprintf("Changes what the data field in JSON outputs will contain.\n\nUse '%s' with CAUTION.", JSONOutputs.Raw),
				}),
			canDefault: true,
			allowed:    JSONOutputs.List(),
		})
	// EnvTOTPFormat supports formatting the TOTP tokens for generation of tokens
	EnvTOTPFormat = environmentRegister(EnvironmentFormatter{environmentBase: environmentBase{
		subKey: "OTP_FORMAT",
		cat:    totpCategory,
		desc:   "Override the otpauth url used to store totp tokens. It must have ONE format string ('%s') to insert the totp base code.",
	}, fxn: formatterTOTP, allowed: "otpauth//url/%s/args..."})
	// EnvConfig is the location of the config file to read
	EnvConfig = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(detectEnvironment,
				environmentBase{
					subKey: "CONFIG_TOML",
					desc: fmt.Sprintf(`Allows setting a specific toml file to read and load into the environment.

The keyword '%s' will disable this functionality and the keyword '%s' will
search for a file in the following paths in XDG_CONFIG_HOME (%s) or from the user's HOME (%s).
Matches the first file found.

Note that this value is not output as part of the environment, nor
can it be set via TOML configuration.`, noEnvironment, detectEnvironment, strings.Join(xdgPaths, ","), strings.Join(homePaths, ",")),
				}),
			canDefault: true,
			allowed:    []string{detectEnvironment, fileExample, noEnvironment},
		})
	// EnvPasswordMode indicates how the password is read
	EnvPasswordMode = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment(string(DefaultKeyMode),
				environmentBase{
					cat:         credsCategory,
					subKey:      "PASSWORD_MODE",
					requirement: "must be set to a valid mode when using a key",
					desc: fmt.Sprintf(`How to retrieve the database store password. Set to '%s' when only using a key file.
Set to '%s' to ignore the set key value`, noKeyMode, IgnoreKeyMode),
				}),
			allowed:    []string{string(AskKeyMode), string(commandKeyMode), string(IgnoreKeyMode), string(noKeyMode), string(plainKeyMode)},
			canDefault: true,
		})
	envPassword = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					cat:         credsCategory,
					requirement: requiredKeyOrKeyFile,
					subKey:      "PASSWORD",
					desc: fmt.Sprintf("The database key ('%s' mode) or command to run ('%s' mode) to retrieve the database password.",
						plainKeyMode,
						commandKeyMode),
				}),
			allowed:    []string{commandArgsExample, "password"},
			canDefault: false,
		})
	// EnvPasswordGenWordCount is the number of words that will be selected for password generation
	EnvPasswordGenWordCount = environmentRegister(
		EnvironmentInt{
			environmentDefault: newDefaultedEnvironment(8,
				environmentBase{
					subKey: "WORD_COUNT",
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
					desc:   fmt.Sprintf("The go text template to use to format the chosen words into a password (use '%s' to include a '$' to avoid shell expansion issues). Available fields: Text, Position.Start, and Position.End.", TemplateVariable),
				}),
			allowed:    []string{"<go template>"},
			canDefault: true,
		})
	// EnvPasswordGenWordList is the command text to generate the word list
	EnvPasswordGenWordList = environmentRegister(EnvironmentCommand{environmentBase: environmentBase{
		subKey: "WORDS_COMMAND",
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
	// EnvPasswordGenEnabled indicates if password generation is enabled
	EnvPasswordGenEnabled = environmentRegister(
		EnvironmentBool{
			environmentDefault: newDefaultedEnvironment(true,
				environmentBase{
					cat:    genCategory,
					subKey: "ENABLED",
					desc:   "Enable password generation.",
				}),
		})
	// EnvPasswordGenChars allows for restricting which characters can be used
	EnvPasswordGenChars = environmentRegister(
		EnvironmentString{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					subKey: "CHARACTERS",
					cat:    genCategory,
					desc:   "The set of allowed characters in output words (empty means any character is allowed).",
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
	val := JSONOutputMode(strings.ToLower(strings.TrimSpace(EnvJSONMode.Get())))
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
		isColored, err := EnvColorEnabled.Get()
		if err != nil {
			return false, err
		}
		colors = isColored
	}
	return colors, nil
}
