package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

type (
	ConfigLoader func(string) (io.Reader, error)
	ShellEnv     struct {
		Key   string
		Value string
	}
)

var (
	//go:embed "config.toml"
	ExampleTOML string
	redirects   = map[string]string{
		"HOOK_DIRECTORY":   EnvHookDir.Key(),
		"HOOK_ENABLED":     EnvNoHooks.Key(),
		"JSON_MODE":        EnvJSONDataOutput.Key(),
		"JSON_HASH_LENGTH": EnvHashLength.Key(),
		"KEYS_FILE":        EnvKeyFile.Key(),
		"KEYS_MODE":        EnvKeyMode.Key(),
		"KEYS_KEY":         envKey.Key(),
		"CLIP_ENABLED":     EnvNoClip.Key(),
		"COLOR_ENABLED":    EnvNoColor.Key(),
		"PWGEN_ENABLED":    EnvNoPasswordGen.Key(),
		"TOTP_ENABLED":     EnvNoTOTP.Key(),
		"TOTP_ATTRIBUTE":   EnvTOTPToken.Key(),
		"ENTRY_MODTIME":    EnvModTime.Key(),
	}
	arrayTypes = []string{
		EnvClipCopy.Key(),
		EnvClipPaste.Key(),
		EnvPasswordGenWordList.Key(),
		envKey.Key(),
		EnvTOTPColorBetween.Key(),
	}
	intTypes = []string{
		EnvClipMax.Key(),
		EnvMaxTOTP.Key(),
		EnvHashLength.Key(),
		EnvPasswordGenCount.Key(),
	}
	boolTypes = []string{
		EnvClipOSC52.Key(),
		EnvNoClip.Key(),
		EnvNoColor.Key(),
		EnvNoHooks.Key(),
		EnvNoPasswordGen.Key(),
		EnvNoTOTP.Key(),
		EnvPasswordGenTitle.Key(),
		EnvReadOnly.Key(),
	}
)

func LoadConfig(r io.Reader, loader ConfigLoader) ([]ShellEnv, error) {
	m := make(map[string]interface{})
	if err := overlayConfig(r, true, &m, loader); err != nil {
		return nil, err
	}
	var allowed []string
	for _, k := range registeredEnv {
		allowed = append(allowed, k.self().Key())
	}
	m = flatten(m, "")
	var res []ShellEnv
	for k, v := range m {
		export := strings.ToUpper(k)
		if r, ok := redirects[export]; ok {
			export = r
		} else {
			export = environmentPrefix + export
		}
		if !slices.Contains(allowed, export) {
			return nil, fmt.Errorf("unknown key: %s (%s)", k, export)
		}
		value, ok := v.(string)
		if !ok {
			if slices.Contains(arrayTypes, export) {
				array, err := parseStringArray(v)
				if err != nil {
					return nil, err
				}
				value = strings.Join(array, " ")
			} else if slices.Contains(intTypes, export) {
				i, ok := v.(int64)
				if !ok {
					return nil, fmt.Errorf("non-int64 found where expected: %v", v)
				}
				if i < 0 {
					return nil, fmt.Errorf("%d is negative (not allowed here)", i)
				}
				value = fmt.Sprintf("%d", i)
			} else if slices.Contains(boolTypes, export) {
				switch t := v.(type) {
				case bool:
					if t {
						value = yes
					} else {
						value = no
					}
				default:
					return nil, fmt.Errorf("non-bool found where expected: %v", v)
				}
			} else {
				return nil, fmt.Errorf("unknown field, can't determine type: %s (%v)", k, v)
			}
		}
		value = os.Expand(value, os.Getenv)
		res = append(res, ShellEnv{Key: export, Value: value})

	}
	return res, nil
}

func overlayConfig(r io.Reader, canInclude bool, m *map[string]interface{}, loader ConfigLoader) error {
	d := toml.NewDecoder(r)
	if _, err := d.Decode(m); err != nil {
		return err
	}
	res := *m
	includes, ok := res["include"]
	if ok {
		delete(*m, "include")
		including, err := parseStringArray(includes)
		if err != nil {
			return err
		}
		if len(including) > 0 {
			if !canInclude {
				return errors.New("nested includes not allowed")
			}
			for _, s := range including {
				read := os.Expand(s, os.Getenv)
				reader, err := loader(read)
				if err != nil {
					return err
				}
				if err := overlayConfig(reader, false, m, nil); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func parseStringArray(value interface{}) ([]string, error) {
	var res []string
	switch t := value.(type) {
	case []interface{}:
		for _, item := range t {
			switch s := item.(type) {
			case string:
				res = append(res, s)
			default:
				return nil, fmt.Errorf("value is not string in array: %v", item)
			}
		}
	default:
		return nil, fmt.Errorf("value is not of array type: %v", value)
	}
	return res, nil
}

func flatten(m map[string]interface{}, prefix string) map[string]interface{} {
	flattened := make(map[string]interface{})

	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "_" + k
		}

		switch to := v.(type) {
		case map[string]interface{}:
			for subKey, subVal := range flatten(to, key) {
				flattened[subKey] = subVal
			}
		default:
			flattened[key] = v
		}
	}

	return flattened
}

func configLoader(path string) (io.Reader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func LoadConfigFile(path string) error {
	reader, err := configLoader(path)
	if err != nil {
		return err
	}
	env, err := LoadConfig(reader, configLoader)
	if err != nil {
		return err
	}
	for _, v := range env {
		os.Setenv(v.Key, v.Value)
	}
	return nil
}
