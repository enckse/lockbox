// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/output"
	"git.sr.ht/~enckse/lockbox/internal/platform"
	"git.sr.ht/~enckse/lockbox/internal/util"
)

var (
	// EnvClipTimeout gets the maximum clipboard time
	EnvClipTimeout = environmentRegister(EnvironmentInt{
		environmentDefault: newDefaultedEnvironment(45,
			environmentBase{
				key:         clipCategory + "TIMEOUT",
				description: "Override the amount of time before totp clears the clipboard (seconds).",
			}),
		short: "clipboard max time",
	})
	// EnvJSONHashLength handles the hashing output length
	EnvJSONHashLength = environmentRegister(EnvironmentInt{
		environmentDefault: newDefaultedEnvironment(0,
			environmentBase{
				key:         jsonCategory + "HASH_LENGTH",
				description: fmt.Sprintf("Maximum string length of the JSON value when '%s' mode is set for JSON output.", output.JSONModes.Hash),
			}),
		short:   "hash length",
		canZero: true,
	})
	// EnvClipOSC52 indicates if OSC52 clipboard mode is enabled
	EnvClipOSC52 = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(false,
			environmentBase{
				key:         clipCategory + "OSC52",
				description: "Enable OSC52 clipboard mode.",
			}),
	})
	// EnvTOTPEnabled indicates if TOTP is allowed
	EnvTOTPEnabled = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         totpCategory + "ENABLED",
				description: "Enable TOTP integrations.",
			}),
	})
	// EnvReadOnly indicates if in read-only mode
	EnvReadOnly = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(false,
			environmentBase{
				key:         "READONLY",
				description: "Operate in readonly mode.",
			}),
	})
	// EnvClipEnabled indicates if clipboard is enabled
	EnvClipEnabled = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         clipCategory + "ENABLED",
				description: "Enable clipboard operations.",
			}),
	})
	// EnvColorEnabled indicates if colors are enabled
	EnvColorEnabled = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         "COLOR_ENABLED",
				description: "Enable terminal colors.",
			}),
	})
	// EnvHooksEnabled indicates if hooks are enabled
	EnvHooksEnabled = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         hookCategory + "ENABLED",
				description: "Enable hooks",
			}),
	})
	// EnvInteractive indicates if operating in interactive mode
	EnvInteractive = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         "INTERACTIVE",
				description: "Enable interactive mode.",
			}),
	})
	// EnvTOTPTimeout indicates when TOTP display should timeout
	EnvTOTPTimeout = environmentRegister(EnvironmentInt{
		environmentDefault: newDefaultedEnvironment(120,
			environmentBase{
				key:         totpCategory + "TIMEOUT",
				description: "Time, in seconds, to show a TOTP token before automatically exiting.",
			}),
		short: "max totp time",
	})
	// EnvTOTPEntry is the leaf token to use to store TOTP tokens
	EnvTOTPEntry = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("totp",
				environmentBase{
					key:         totpCategory + "ENTRY",
					description: "Entry name to store TOTP tokens within the database.",
				}),
			allowed: []string{"<string>"},
			flags:   []stringsFlags{canDefaultFlag},
		},
	})
	// EnvPlatform is the platform that the application is running on
	EnvPlatform = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment(detectedValue,
				environmentBase{
					key:         "PLATFORM",
					description: "Override the detected platform.",
				}),
			allowed: platform.Systems.List(),
		},
	})
	// EnvStore is the location of the keepass file/store
	EnvStore = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         "STORE",
					description: "The kdbx file to operate on.",
					requirement: "must be set",
				}),
			allowed: []string{fileExample},
			flags:   []stringsFlags{canExpandFlag},
		},
	})
	// EnvHookDir is the directory of hooks to execute
	EnvHookDir = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         hookCategory + "DIRECTORY",
					description: "The path to a directory of hooks to execute on actions against the database.",
				}),
			allowed: []string{"<directory>"},
			flags:   []stringsFlags{canDefaultFlag, canExpandFlag},
		},
	})
	// EnvClipCopy allows overriding the clipboard copy command
	EnvClipCopy = environmentRegister(EnvironmentArray{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         clipCategory + "COPY_COMMAND",
					description: "Override the detected platform copy command.",
				}),
			flags: []stringsFlags{isCommandFlag},
		},
	})
	// EnvClipPaste allows overriding the clipboard paste command
	EnvClipPaste = environmentRegister(EnvironmentArray{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         clipCategory + "PASTE_COMMAND",
					description: "Override the detected platform paste command.",
				}),
			flags: []stringsFlags{isCommandFlag},
		},
	})
	// EnvTOTPColorBetween handles terminal coloring for TOTP windows (seconds)
	EnvTOTPColorBetween = environmentRegister(EnvironmentArray{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment(strings.Join(TOTPDefaultBetween, arrayDelimiter),
				environmentBase{
					key: totpCategory + "COLOR_WINDOWS",
					description: fmt.Sprintf(`Override when to set totp generated outputs to different colors,
must be a list of one (or more) rules where a '%s' delimits the start and end second (0-60 for each).`, util.TimeWindowSpan),
				}),
			flags:   []stringsFlags{canDefaultFlag},
			allowed: exampleColorWindows,
		},
	})
	// EnvKeyFile is an keyfile for the database
	EnvKeyFile = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         credsCategory + "KEY_FILE",
					requirement: requiredKeyOrKeyFile,
					description: "A keyfile to access/protect the database.",
				}),
			allowed: []string{"keyfile"},
			flags:   []stringsFlags{canDefaultFlag, canExpandFlag},
		},
	})
	// EnvDefaultModTime is modtime override ability for entries
	EnvDefaultModTime = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         defaultCategory + "MODTIME",
					description: fmt.Sprintf("Input modification time to set for the entry\n\nExpected format: %s.", ModTimeFormat),
				}),
			flags:   []stringsFlags{canDefaultFlag},
			allowed: []string{"modtime"},
		},
	})
	// EnvJSONMode controls how JSON is output in the 'data' field
	EnvJSONMode = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment(string(output.JSONModes.Hash),
				environmentBase{
					key:         jsonCategory + "MODE",
					description: fmt.Sprintf("Changes what the data field in JSON outputs will contain.\n\nUse '%s' with CAUTION.", output.JSONModes.Raw),
				}),
			flags:   []stringsFlags{canDefaultFlag},
			allowed: output.JSONModes.List(),
		},
	})
	// EnvTOTPFormat supports formatting the TOTP tokens for generation of tokens
	EnvTOTPFormat = environmentRegister(EnvironmentFormatter{
		environmentBase: environmentBase{
			key:         totpCategory + "OTP_FORMAT",
			description: "Override the otpauth url used to store totp tokens. It must have ONE format string ('%s') to insert the totp base code.",
		}, fxn: formatterTOTP, allowed: "otpauth//url/%s/args...",
	})
	// EnvPasswordMode indicates how the password is read
	EnvPasswordMode = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment(string(DefaultKeyMode),
				environmentBase{
					key:         credsCategory + "PASSWORD_MODE",
					requirement: "must be set to a valid mode when using a key",
					description: fmt.Sprintf(`How to retrieve the database store password. Set to '%s' when only using a key file.
Set to '%s' to ignore the set key value`, noKeyMode, IgnoreKeyMode),
				}),
			allowed: []string{string(AskKeyMode), string(commandKeyMode), string(IgnoreKeyMode), string(noKeyMode), string(plainKeyMode)},
			flags:   []stringsFlags{canDefaultFlag},
		},
	})
	envPassword = environmentRegister(EnvironmentArray{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment(unset,
				environmentBase{
					requirement: requiredKeyOrKeyFile,
					key:         credsCategory + "PASSWORD",
					description: fmt.Sprintf("The database password itself ('%s' mode) or command to run ('%s' mode) to retrieve the database password.",
						plainKeyMode,
						commandKeyMode),
				}),
			allowed: []string{commandArgsExample, "password"},
			flags:   []stringsFlags{canExpandFlag},
		},
	})
	// EnvPasswordGenWordCount is the number of words that will be selected for password generation
	EnvPasswordGenWordCount = environmentRegister(EnvironmentInt{
		environmentDefault: newDefaultedEnvironment(8,
			environmentBase{
				key:         genCategory + "WORD_COUNT",
				description: "Number of words to select and include in the generated password.",
			}),
		short: "word count",
	})
	// EnvPasswordGenTemplate is the output template for controlling how output words are placed together
	EnvPasswordGenTemplate = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("{{range $i, $val := .}}{{if $i}}-{{end}}{{$val.Text}}{{end}}",
				environmentBase{
					key:         genCategory + "TEMPLATE",
					description: fmt.Sprintf("The go text template to use to format the chosen words into a password. Available fields: %s.", util.TextPositionFields()),
				}),
			allowed: []string{"<go template>"},
			flags:   []stringsFlags{canDefaultFlag},
		},
	})
	// EnvPasswordGenWordList is the command text to generate the word list
	EnvPasswordGenWordList = environmentRegister(EnvironmentArray{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         genCategory + "WORD_COMMAND",
					description: "Command to retrieve the word list to use for password generation (must be split by newline).",
				}),
			flags: []stringsFlags{isCommandFlag},
		},
	})
	// EnvPasswordGenEnabled indicates if password generation is enabled
	EnvPasswordGenEnabled = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         genCategory + "ENABLED",
				description: "Enable password generation.",
			}),
	})
	// EnvPasswordGenChars allows for restricting which characters can be used
	EnvPasswordGenChars = environmentRegister(EnvironmentString{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         genCategory + "CHARACTERS",
					description: "The set of allowed characters in output words (empty means any character is allowed).",
				}),
			allowed: []string{"<list of characters>"},
			flags:   []stringsFlags{canDefaultFlag},
		},
	})
)
