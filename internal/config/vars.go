// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"strings"

	"github.com/enckse/lockbox/internal/output"
)

var (
	// EnvFeatureTOTP allows disable TOTP feature
	EnvFeatureTOTP = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         featureCategory + "TOTP",
				description: "Enable totp feature.",
			}),
	})
	// EnvFeatureClip allows disabling clipboard feature
	EnvFeatureClip = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         featureCategory + "CLIP",
				description: "Enable clipboard feature.",
			}),
	})
	// EnvFeatureColor allows disabling color output
	EnvFeatureColor = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         featureCategory + "COLOR",
				description: "Enable terminal color feature.",
			}),
	})
	// EnvJSONHashLength handles the hashing output length
	EnvJSONHashLength = environmentRegister(EnvironmentInt{
		environmentDefault: newDefaultedEnvironment(1,
			environmentBase{
				key: jsonCategory + "HASH_LENGTH",
				description: fmt.Sprintf(`Maximum string length of the JSON checksum value when JSON '%s' mode is set.
This value is appended to the single character type field`, output.JSONModes.Hash),
			}),
		short: "checksum value length",
	})
	// EnvReadOnly indicates if in read-only mode
	EnvReadOnly = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(false,
			environmentBase{
				key:         "READONLY",
				description: "Operate in readonly mode.",
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
	// EnvTOTPCheckOnInsert will indicate if TOTP tokens should be check for validity during the insert process
	EnvTOTPCheckOnInsert = environmentRegister(EnvironmentBool{
		environmentDefault: newDefaultedEnvironment(true,
			environmentBase{
				key:         totpCategory + "CHECK_ON_INSERT",
				description: "Test TOTP code generation on insert.",
			}),
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
	// EnvClipCopy allows overriding the clipboard copy command
	EnvClipCopy = environmentRegister(EnvironmentArray{
		environmentStrings: environmentStrings{
			environmentDefault: newDefaultedEnvironment("",
				environmentBase{
					key:         clipCategory + "COPY",
					description: "Override the detected platform copy command.",
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
must be a list of one (or more) rules where a '%s' delimits the start and end second (0-60 for each).`, TimeWindowSpan),
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
			allowed: []string{string(commandKeyMode), string(IgnoreKeyMode), string(noKeyMode), string(plainKeyMode)},
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
)
