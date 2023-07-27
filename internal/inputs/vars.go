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
	prefixKey            = "LOCKBOX_"
	clipBaseEnv          = prefixKey + "CLIP_"
	plainKeyMode         = "plaintext"
	commandKeyMode       = "command"
	commandArgsExample   = "[cmd args...]"
	detectedValue        = "(detected)"
	requiredKeyOrKeyFile = "a key, a key file, or both must be set"
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
	// EnvClipMax gets the maximum clipboard time
	EnvClipMax = EnvironmentInt{environmentBase: environmentBase{key: clipBaseEnv + "MAX", desc: "override the amount of time before totp clears the clipboard (e.g. 10),\nmust be an integer"}, shortDesc: "clipboard max time", allowZero: false, defaultValue: 45}
	// EnvHashLength handles the hashing output length
	EnvHashLength = EnvironmentInt{environmentBase: environmentBase{key: EnvJSONDataOutput.key + "_HASH_LENGTH", desc: fmt.Sprintf("maximum hash length the JSON output should contain\nwhen '%s' mode is set for JSON output", JSONDataOutputHash)}, shortDesc: "hash length", allowZero: true, defaultValue: 0}
	// EnvClipOSC52 indicates if OSC52 clipboard mode is enabled
	EnvClipOSC52 = EnvironmentBool{environmentBase: environmentBase{key: clipBaseEnv + "OSC52", desc: "enable OSC52 clipboard mode"}, defaultValue: false}
	// EnvNoTOTP indicates if TOTP is disabled
	EnvNoTOTP = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "NOTOTP", desc: "disable TOTP integrations"}, defaultValue: false}
	// EnvReadOnly indicates if in read-only mode
	EnvReadOnly = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "READONLY", desc: "operate in readonly mode"}, defaultValue: false}
	// EnvNoClip indicates clipboard functionality is off
	EnvNoClip = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "NOCLIP", desc: "disable clipboard operations"}, defaultValue: false}
	// EnvNoColor indicates if color outputs are disabled
	EnvNoColor = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "NOCOLOR", desc: "disable terminal colors"}, defaultValue: false}
	// EnvInteractive indicates if operating in interactive mode
	EnvInteractive = EnvironmentBool{environmentBase: environmentBase{key: prefixKey + "INTERACTIVE", desc: "enable interactive mode"}, defaultValue: true}
	// EnvMaxTOTP is the max TOTP time to run (default)
	EnvMaxTOTP = EnvironmentInt{environmentBase: environmentBase{key: EnvTOTPToken.key + "_MAX", desc: "time, in seconds, in which to show a TOTP token before automatically exiting"}, shortDesc: "max totp time", allowZero: false, defaultValue: 120}
	// EnvTOTPToken is the leaf token to use to store TOTP tokens
	EnvTOTPToken = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "TOTP", desc: "attribute name to store TOTP tokens within the database"}, allowed: []string{"string"}, canDefault: true, defaultValue: "totp"}
	// EnvPlatform is the platform that the application is running on
	EnvPlatform = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "PLATFORM", desc: "override the detected platform"}, defaultValue: detectedValue, allowed: PlatformSet(), canDefault: false}
	// EnvStore is the location of the keepass file/store
	EnvStore = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "STORE", desc: "directory to the database file", requirement: "must be set"}, canDefault: false, allowed: []string{"file"}}
	// EnvHookDir is the directory of hooks to execute
	EnvHookDir = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "HOOKDIR", desc: "the path to hooks to execute on actions against the database"}, allowed: []string{"directory"}, canDefault: true, defaultValue: ""}
	// EnvClipCopy allows overriding the clipboard copy command
	EnvClipCopy = EnvironmentCommand{environmentBase: environmentBase{key: clipBaseEnv + "COPY", desc: "override the detected platform copy command"}}
	// EnvClipPaste allows overriding the clipboard paste command
	EnvClipPaste = EnvironmentCommand{environmentBase: environmentBase{key: clipBaseEnv + "PASTE", desc: "override the detected platform paste command"}}
	// EnvTOTPColorBetween handles terminal coloring for TOTP windows (seconds)
	EnvTOTPColorBetween = EnvironmentString{environmentBase: environmentBase{key: EnvTOTPToken.key + "_BETWEEN", desc: "override when to set totp generated outputs to different colors, must be a\nlist of one (or more) rules where a semicolon delimits the start and end\nsecond (0-60 for each)"}, canDefault: true, defaultValue: TOTPDefaultBetween, allowed: []string{"start:end,start:end,start:end..."}}
	// EnvKeyFile is an keyfile for the database
	EnvKeyFile = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "KEYFILE", requirement: requiredKeyOrKeyFile, desc: "keyfile to access/protect the database"}, allowed: []string{"keyfile"}, canDefault: true, defaultValue: ""}
	// EnvModTime is modtime override ability for entries
	EnvModTime = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "SET_MODTIME", desc: fmt.Sprintf("input modification time to set for the entry\n(expected format: %s)", ModTimeFormat)}, canDefault: true, defaultValue: "", allowed: []string{"modtime"}}
	// EnvJSONDataOutput controls how JSON is output in the 'data' field
	EnvJSONDataOutput = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "JSON_DATA_OUTPUT", desc: fmt.Sprintf("changes what the data field in JSON outputs will contain\nuse '%s' with CAUTION", JSONDataOutputRaw)}, canDefault: true, defaultValue: string(JSONDataOutputHash), allowed: []string{string(JSONDataOutputRaw), string(JSONDataOutputHash), string(JSONDataOutputBlank)}}
	// EnvFormatTOTP supports formatting the TOTP tokens for generation of tokens
	EnvFormatTOTP = EnvironmentFormatter{environmentBase: environmentBase{key: EnvTOTPToken.key + "_FORMAT", desc: "override the otpauth url used to store totp tokens. It must have ONE format\nstring ('%s') to insert the totp base code"}, fxn: formatterTOTP, allowed: "otpauth//url/%s/args..."}
	envKeyMode    = EnvironmentString{environmentBase: environmentBase{key: prefixKey + "KEYMODE", requirement: "must be set to a valid mode when using a key", desc: "how to retrieve the database store password"}, allowed: []string{commandKeyMode, plainKeyMode}, canDefault: true, defaultValue: commandKeyMode}
	envKey        = EnvironmentString{environmentBase: environmentBase{requirement: requiredKeyOrKeyFile, key: prefixKey + "KEY", desc: fmt.Sprintf("the database key ('%s' mode) or command to run ('%s' mode)\nto retrieve the database password", plainKeyMode, commandKeyMode)}, allowed: []string{commandArgsExample, "password"}, canDefault: false}
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
	type keyer struct {
		env EnvironmentString
		has bool
		in  string
	}
	check := func(in string, e EnvironmentString) keyer {
		val := strings.TrimSpace(in)
		return keyer{has: val != "", env: e, in: in}
	}
	inStore := check(*store, EnvStore)
	inKey := check(*key, envKey)
	inKeyFile := check(*keyFile, EnvKeyFile)
	inKeyMode := check(*keyMode, envKeyMode)
	var out []string
	for _, k := range []keyer{inStore, inKey, inKeyFile, inKeyMode} {
		out = append(out, k.env.KeyValue(k.in))
	}
	sort.Strings(out)
	if !inStore.has || (!inKey.has && !inKeyFile.has) {
		return nil, fmt.Errorf("missing required arguments for rekey: %s", strings.Join(out, " "))
	}
	return out, nil
}

// GetKey will get the encryption key setup for lb
func GetKey() ([]byte, error) {
	useKey := envKey.Get()
	if useKey == "" {
		return nil, nil
	}
	var data []byte
	switch envKeyMode.Get() {
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

// ListEnvironmentVariables will print information about env variables and potential/set values
func ListEnvironmentVariables(showValues bool) []string {
	out := environmentOutput{showValues: showValues}
	var results []string
	for _, item := range []printer{EnvStore, envKeyMode, envKey, EnvNoClip, EnvNoColor, EnvInteractive, EnvReadOnly, EnvTOTPToken, EnvFormatTOTP, EnvMaxTOTP, EnvTOTPColorBetween, EnvClipPaste, EnvClipCopy, EnvClipMax, EnvPlatform, EnvNoTOTP, EnvHookDir, EnvClipOSC52, EnvKeyFile, EnvModTime, EnvJSONDataOutput, EnvHashLength} {
		env := item.self()
		value, allow := item.values()
		if out.showValues {
			value = os.Getenv(env.key)
		}
		if len(value) == 0 {
			value = "(unset)"
		}
		description := strings.ReplaceAll(env.desc, "\n", "\n  ")
		requirement := "optional/default"
		r := strings.TrimSpace(env.requirement)
		if r != "" {
			requirement = r
		}
		text := fmt.Sprintf("\n%s\n  %s\n\n  requirement: %s\n  value: %s\n  options: %s\n", env.key, description, requirement, value, strings.Join(allow, "|"))
		results = append(results, text)
	}
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
