package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	isInclude = "include"
	maxDepth  = 10
)

type (
	// Loader indicates how included files should be sourced
	Loader func(string) (io.Reader, error)
	// ShellEnv is the output shell environment settings parsed from TOML config
	ShellEnv struct {
		Key   string
		Value string
		raw   string
	}
)

var (
	// ExampleTOML is an example TOML file of viable fields
	//go:embed "config.toml"
	ExampleTOML string
	arrayTypes  = []string{
		EnvClipCopy.Key(),
		EnvClipPaste.Key(),
		EnvPasswordGenWordList.Key(),
		envPassword.Key(),
		EnvTOTPColorBetween.Key(),
	}
	intTypes = []string{
		EnvClipTimeout.Key(),
		EnvTOTPTimeout.Key(),
		EnvJSONHashLength.Key(),
		EnvPasswordGenWordCount.Key(),
	}
	boolTypes = []string{
		EnvClipOSC52.Key(),
		EnvClipEnabled.Key(),
		EnvTOTPEnabled.Key(),
		EnvColorEnabled.Key(),
		EnvHooksEnabled.Key(),
		EnvPasswordGenEnabled.Key(),
		EnvPasswordGenTitle.Key(),
		EnvReadOnly.Key(),
		EnvInteractive.Key(),
		EnvDefaultCompletion.Key(),
	}
	reverseMap = map[string][]string{
		"[]":   arrayTypes,
		"0":    intTypes,
		"true": boolTypes,
	}
)

// DefaultTOML will load the internal, default TOML with additional comment markups
func DefaultTOML() (string, error) {
	s, err := LoadConfig(strings.NewReader(ExampleTOML), nil)
	if err != nil {
		return "", err
	}
	const root = "_root_"
	unmapped := make(map[string][]string)
	keys := []string{}
	for _, item := range s {
		raw := item.raw
		parts := strings.Split(raw, "_")
		length := len(parts)
		if length == 0 {
			return "", fmt.Errorf("invalid internal TOML structure: %v", item)
		}
		key := parts[0]
		sub := ""
		switch length {
		case 1:
			key = root
			sub = parts[0]
		case 2:
			sub = parts[1]
		default:
			sub = strings.Join(parts[1:], "_")
		}
		field := "\"\""
		for to, fromKey := range reverseMap {
			if slices.Contains(fromKey, item.Key) {
				field = to
				break
			}
		}
		text, err := generateDetailText(item.Key)
		if err != nil {
			return "", err
		}
		sub = fmt.Sprintf(`%s
%s = %s
`, text, sub, field)
		had, ok := unmapped[key]
		if !ok {
			had = []string{}
			keys = append(keys, key)
		}
		had = append(had, sub)
		unmapped[key] = had
	}
	sort.Strings(keys)
	builder := strings.Builder{}
	configEnv, err := generateDetailText(EnvConfig.Key())
	if err != nil {
		return "", err
	}
	for _, header := range []string{configEnv, "\n", fmt.Sprintf(`
# include additional configs, allowing globs ('*'), nesting
# depth allowed up to %d include levels
#
# this field is not configurable via environment variables
# and it is not considered part of the environment either
# it is ONLY used during TOML configuration loading
%s = []
`, maxDepth, isInclude), "\n"} {
		if _, err := builder.WriteString(header); err != nil {
			return "", err
		}
	}
	for _, k := range keys {
		if k != root {
			if _, err := fmt.Fprintf(&builder, "\n[%s]\n", k); err != nil {
				return "", err
			}
		}
		for _, sub := range unmapped[k] {
			if _, err := builder.WriteString(sub); err != nil {
				return "", err
			}
		}
	}
	return builder.String(), nil
}

func generateDetailText(key string) (string, error) {
	data, ok := registry[key]
	if !ok {
		return "", fmt.Errorf("unexpected configuration key has no environment settings: %s", key)
	}
	env := data.self()
	value, allow := data.values()
	if len(value) == 0 {
		value = "(unset)"
	}
	description := strings.TrimSpace(Wrap(2, env.desc))
	requirement := "optional/default"
	r := strings.TrimSpace(env.requirement)
	if r != "" {
		requirement = r
	}
	var text []string
	for _, line := range []string{fmt.Sprintf("environment: %s", key), fmt.Sprintf("description:\n%s\n", description), fmt.Sprintf("default: %s", requirement), fmt.Sprintf("option: %s", strings.Join(allow, "|")), fmt.Sprintf("default: %s", value)} {
		for _, comment := range strings.Split(line, "\n") {
			text = append(text, fmt.Sprintf("# %s", comment))
		}
	}
	return strings.Join(text, "\n"), nil
}

// LoadConfig will read the input reader and use the loader to source configuration files
func LoadConfig(r io.Reader, loader Loader) ([]ShellEnv, error) {
	maps, err := readConfigs(r, 1, loader)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	for _, config := range maps {
		for k, v := range flatten(config, "") {
			m[k] = v
		}
	}
	var res []ShellEnv
	for k, v := range m {
		export := environmentPrefix + strings.ToUpper(k)
		if _, ok := registry[export]; !ok {
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
		res = append(res, ShellEnv{Key: export, Value: value, raw: k})

	}
	return res, nil
}

func readConfigs(r io.Reader, depth int, loader Loader) ([]map[string]interface{}, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("too many nested includes (%d > %d)", depth, maxDepth)
	}
	d := toml.NewDecoder(r)
	m := make(map[string]interface{})
	if _, err := d.Decode(&m); err != nil {
		return nil, err
	}
	maps := []map[string]interface{}{m}
	includes, ok := m[isInclude]
	if ok {
		delete(m, isInclude)
		including, err := parseStringArray(includes)
		if err != nil {
			return nil, err
		}
		if len(including) > 0 {
			for _, s := range including {
				use := os.Expand(s, os.Getenv)
				files := []string{use}
				if strings.Contains(use, "*") {
					matched, err := filepath.Glob(use)
					if err != nil {
						return nil, err
					}
					files = matched
				}
				for _, file := range files {
					reader, err := loader(file)
					if err != nil {
						return nil, err
					}
					results, err := readConfigs(reader, depth+1, loader)
					if err != nil {
						return nil, err
					}
					maps = append(maps, results...)
				}
			}
		}
	}
	return maps, nil
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

// LoadConfigFile will load a path as the configuration
// it will also set the environment
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
