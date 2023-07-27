// Package inputs handles user inputs/UI elements.
package inputs

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	prefixKey     = "LOCKBOX_"
	fieldTOTPEnv  = prefixKey + "TOTP"
	clipBaseEnv   = prefixKey + "CLIP_"
	formatTOTPEnv = fieldTOTPEnv + "_FORMAT"
	keyModeEnv    = prefixKey + "KEYMODE"
	keyEnv        = prefixKey + "KEY"
	// KeyFileEnv is an OPTIONAL keyfile for the database
	KeyFileEnv         = prefixKey + "KEYFILE"
	plainKeyMode       = "plaintext"
	commandKeyMode     = "command"
	defaultTOTPField   = "totp"
	commandArgsExample = "[cmd args...]"
	detectedValue      = "(detected)"
	// ModTimeEnv is modtime override ability for entries
	ModTimeEnv = prefixKey + "SET_MODTIME"
	// ModTimeFormat is the expected modtime format
	ModTimeFormat = time.RFC3339
	// JSONDataOutputEnv controls how JSON is output
	JSONDataOutputEnv = prefixKey + "JSON_DATA_OUTPUT"
	// JSONDataOutputHash means output data is hashed
	JSONDataOutputHash JSONOutputMode = "hash"
	// JSONDataOutputBlank means an empty entry is set
	JSONDataOutputBlank JSONOutputMode = "empty"
	// JSONDataOutputRaw means the RAW (unencrypted) value is displayed
	JSONDataOutputRaw JSONOutputMode = "plaintext"
)

var (
	// EnvClipboardMax gets the maximum clipboard time
	EnvClipboardMax = EnvironmentInt{environmentBase: environmentBase{key: clipBaseEnv + "MAX"}, shortDesc: "clipboard max time", allowZero: false, defaultValue: 45}
	// EnvHashLength handles the hashing output length
	EnvHashLength = EnvironmentInt{environmentBase: environmentBase{key: JSONDataOutputEnv + "_HASH_LENGTH"}, shortDesc: "hash length", allowZero: true, defaultValue: 0}
	// EnvClipOSC52 indicates if OSC52 clipboard mode is enabled
	EnvClipOSC52 = EnvironmentBool{environmentBase: environmentBase{key: clipBaseEnv + "OSC52"}, defaultValue: false}
	// EnvNoTOTP indicates if TOTP is disabled
	EnvNoTOTP = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "NOTOTP"}, defaultValue: false}
	// EnvReadOnly indicates if in read-only mode
	EnvReadOnly = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "READONLY"}, defaultValue: false}
	// EnvNoClip indicates clipboard functionality is off
	EnvNoClip = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "NOCLIP"}, defaultValue: false}
	// EnvNoColor indicates if color outputs are disabled
	EnvNoColor = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "NOCOLOR"}, defaultValue: false}
	// EnvInteractive indicates if operating in interactive mode
	EnvInteractive = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "INTERACTIVE"}, defaultValue: true}
	// EnvMaxTOTP is the max TOTP time to run (default)
	EnvMaxTOTP = EnvironmentInt{environmentBase: environmentBase{key: fieldTOTPEnv + "_MAX"}, shortDesc: "max totp time", allowZero: false, defaultValue: 120}
	// EnvTOTPToken is the leaf token to use to store TOTP tokens
	EnvTOTPToken = EnvironmentString{environmentBase: environmentBase{key: fieldTOTPEnv}, canDefault: true, defaultValue: "totp"}
	// EnvPlatform is the platform that the application is running on
	EnvPlatform = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "PLATFORM"}, canDefault: false}
	// EnvStore is the location of the keepass file/store
	EnvStore = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "STORE"}, canDefault: false}
	// EnvHookDir is the directory of hooks to execute
	EnvHookDir = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "HOOKDIR"}, canDefault: true, defaultValue: ""}
	// EnvClipCopy allows overriding the clipboard copy command
	EnvClipCopy = EnvironmentCommand{environmentBase: environmentBase{key: clipBaseEnv + "COPY"}}
	// EnvClipPaste allows overriding the clipboard paste command
	EnvClipPaste = EnvironmentCommand{environmentBase: environmentBase{key: clipBaseEnv + "PASTE"}}
	// EnvColorBetween handles terminal coloring for TOTP windows (seconds)
	EnvColorBetween = EnvironmentString{environmentBase: environmentBase{key: fieldTOTPEnv + "_BETWEEN"}, canDefault: true, defaultValue: TOTPDefaultBetween}
)

type (
	// JSONOutputMode is the output mode definition
	JSONOutputMode    string
	environmentOutput struct {
		showValues bool
	}
)

// GetReKey will get the rekey environment settings
func GetReKey(args []string) ([]string, error) {
	set := flag.NewFlagSet("rekey", flag.ExitOnError)
	store := set.String("store", "", "new store")
	key := set.String("key", "", "new key")
	keyFile := set.String("keyfile", "", "new keyfile")
	keyMode := set.String("keymode", "", "new keymode")
	if err := set.Parse(args); err != nil {
		return nil, err
	}
	mapped := map[string]string{
		keyModeEnv:   *keyMode,
		keyEnv:       *key,
		KeyFileEnv:   *keyFile,
		EnvStore.key: *store,
	}
	hasStore := false
	hasKey := false
	hasKeyFile := false
	var out []string
	for k, val := range mapped {
		if val != "" {
			switch k {
			case EnvStore.key:
				hasStore = true
			case keyEnv:
				hasKey = true
			case KeyFileEnv:
				hasKeyFile = true
			}
		}
		out = append(out, fmt.Sprintf("%s=%s", k, val))
	}
	sort.Strings(out)
	if !hasStore || (!hasKey && !hasKeyFile) {
		return nil, fmt.Errorf("missing required arguments for rekey: %s", strings.Join(out, " "))
	}
	return out, nil
}

// GetKey will get the encryption key setup for lb
func GetKey() ([]byte, error) {
	useKeyMode := os.Getenv(keyModeEnv)
	if useKeyMode == "" {
		useKeyMode = commandKeyMode
	}
	useKey := os.Getenv(keyEnv)
	if useKey == "" {
		return nil, errors.New("no key given")
	}
	var data []byte
	switch useKeyMode {
	case commandKeyMode:
		parts, err := shlex(useKey)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		b, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		data = b
	case plainKeyMode:
		data = []byte(useKey)
	default:
		return nil, errors.New("unknown keymode")
	}
	b := []byte(strings.TrimSpace(string(data)))
	if len(b) == 0 {
		return nil, errors.New("key is empty")
	}
	return b, nil
}

func (o environmentOutput) formatEnvironmentVariable(required bool, name, val, desc string, allowed []string) string {
	value := val
	if o.showValues {
		value = os.Getenv(name)
	}
	if len(value) == 0 {
		value = "(unset)"
	}
	description := strings.ReplaceAll(desc, "\n", "\n  ")
	return fmt.Sprintf("\n%s\n  %s\n\n  required: %t\n  value: %s\n  options: %s\n", name, description, required, value, strings.Join(allowed, "|"))
}

// ListEnvironmentVariables will print information about env variables and potential/set values
func ListEnvironmentVariables(showValues bool) []string {
	e := environmentOutput{showValues: showValues}
	var results []string
	results = append(results, e.formatEnvironmentVariable(true, EnvStore.key, "", "directory to the database file", []string{"file"}))
	results = append(results, e.formatEnvironmentVariable(true, keyModeEnv, commandKeyMode, "how to retrieve the database store password", []string{commandKeyMode, plainKeyMode}))
	results = append(results, e.formatEnvironmentVariable(true, keyEnv, "", fmt.Sprintf("the database key ('%s' mode) or command to run ('%s' mode)\nto retrieve the database password", plainKeyMode, commandKeyMode), []string{commandArgsExample, "password"}))
	results = append(results, e.formatEnvironmentVariable(false, EnvNoClip.key, no, "disable clipboard operations", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, EnvNoColor.key, no, "disable terminal colors", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, EnvInteractive.key, yes, "enable interactive mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, EnvReadOnly.key, no, "operate in readonly mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, fieldTOTPEnv, defaultTOTPField, "attribute name to store TOTP tokens within the database", []string{"string"}))
	results = append(results, e.formatEnvironmentVariable(false, formatTOTPEnv, strings.ReplaceAll(strings.ReplaceAll(FormatTOTP("%s"), "%25s", "%s"), "&", " \\\n           &"), "override the otpauth url used to store totp tokens. It must have ONE format\nstring ('%s') to insert the totp base code", []string{"otpauth//url/%s/args..."}))
	results = append(results, e.formatEnvironmentVariable(false, EnvMaxTOTP.key, fmt.Sprintf("%d", EnvMaxTOTP.defaultValue), "time, in seconds, in which to show a TOTP token before automatically exiting", intArgs))
	results = append(results, e.formatEnvironmentVariable(false, EnvColorBetween.key, TOTPDefaultBetween, "override when to set totp generated outputs to different colors, must be a\nlist of one (or more) rules where a semicolon delimits the start and end\nsecond (0-60 for each)", []string{"start:end,start:end,start:end..."}))
	results = append(results, e.formatEnvironmentVariable(false, EnvClipPaste.key, detectedValue, "override the detected platform paste command", []string{commandArgsExample}))
	results = append(results, e.formatEnvironmentVariable(false, EnvClipCopy.key, detectedValue, "override the detected platform copy command", []string{commandArgsExample}))
	results = append(results, e.formatEnvironmentVariable(false, EnvClipboardMax.key, fmt.Sprintf("%d", EnvClipboardMax.defaultValue), "override the amount of time before totp clears the clipboard (e.g. 10),\nmust be an integer", intArgs))
	results = append(results, e.formatEnvironmentVariable(false, EnvPlatform.key, detectedValue, "override the detected platform", PlatformSet()))
	results = append(results, e.formatEnvironmentVariable(false, EnvNoTOTP.key, no, "disable TOTP integrations", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, EnvHookDir.key, "", "the path to hooks to execute on actions against the database", []string{"directory"}))
	results = append(results, e.formatEnvironmentVariable(false, EnvClipOSC52.key, no, "enable OSC52 clipboard mode", isYesNoArgs))
	results = append(results, e.formatEnvironmentVariable(false, KeyFileEnv, "", "additional keyfile to access/protect the database", []string{"keyfile"}))
	results = append(results, e.formatEnvironmentVariable(false, ModTimeEnv, ModTimeFormat, fmt.Sprintf("input modification time to set for the entry\n(expected format: %s)", ModTimeFormat), []string{"modtime"}))
	results = append(results, e.formatEnvironmentVariable(false, JSONDataOutputEnv, string(JSONDataOutputHash), fmt.Sprintf("changes what the data field in JSON outputs will contain\nuse '%s' with CAUTION", JSONDataOutputRaw), []string{string(JSONDataOutputRaw), string(JSONDataOutputHash), string(JSONDataOutputBlank)}))
	results = append(results, e.formatEnvironmentVariable(false, EnvHashLength.key, fmt.Sprintf("%d", EnvHashLength.defaultValue), fmt.Sprintf("maximum hash length the JSON output should contain\nwhen '%s' mode is set for JSON output", JSONDataOutputHash), intArgs))
	return results
}

// FormatTOTP will format a totp otpauth url
func FormatTOTP(value string) string {
	const (
		otpAuth   = "otpauth"
		otpIssuer = "lbissuer"
	)
	if strings.HasPrefix(value, otpAuth) {
		return value
	}
	override := EnvironOrDefault(formatTOTPEnv, "")
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
	val := strings.ToLower(strings.TrimSpace(EnvironOrDefault(JSONDataOutputEnv, string(JSONDataOutputHash))))
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
