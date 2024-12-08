package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/seanenck/lockbox/internal/app/commands"
	"github.com/seanenck/lockbox/internal/config/store"
	"github.com/seanenck/lockbox/internal/util"
)

const (
	isInclude  = "include"
	maxDepth   = 10
	tomlInt    = "integer"
	tomlBool   = "boolean"
	tomlString = "string"
	tomlArray  = "[]string"
)

type (
	tomlType string
	// Loader indicates how included files should be sourced
	Loader func(string) (io.Reader, error)
)

// DefaultTOML will load the internal, default TOML with additional comment markups
func DefaultTOML() (string, error) {
	const root = "_root_"
	unmapped := make(map[string][]string)
	keys := []string{}
	for envKey, item := range registry {
		tomlKey := strings.ToLower(strings.TrimPrefix(envKey, environmentPrefix))
		parts := strings.Split(tomlKey, "_")
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
		_, field := item.toml()
		text, err := generateDetailText(item)
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
	for _, header := range []string{fmt.Sprintf(`
# include additional configs, allowing globs ('*'), nesting
# depth allowed up to %d include levels
#
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
		subs := unmapped[k]
		sort.Strings(subs)
		for _, sub := range subs {
			if _, err := builder.WriteString(sub); err != nil {
				return "", err
			}
		}
	}
	return builder.String(), nil
}

func generateDetailText(data printer) (string, error) {
	env := data.self()
	value, allow := data.values()
	if len(value) == 0 {
		value = "(unset)"
	}
	key := env.Key()
	description := strings.TrimSpace(util.TextWrap(2, env.desc))
	requirement := "optional/default"
	r := strings.TrimSpace(env.requirement)
	if r != "" {
		requirement = r
	}
	t, _ := data.toml()
	var text []string
	for _, line := range []string{
		fmt.Sprintf("description:\n%s\n", description),
		fmt.Sprintf("requirement: %s", requirement),
		fmt.Sprintf("option: %s", strings.Join(allow, "|")),
		fmt.Sprintf("%s name: %s", commands.Env, key),
		fmt.Sprintf("default: %s", value),
		fmt.Sprintf("type: %s", t),
		"",
		"NOTE: the following value is NOT a default, it is an empty TOML placeholder",
	} {
		for _, comment := range strings.Split(line, "\n") {
			text = append(text, fmt.Sprintf("# %s", comment))
		}
	}
	return strings.Join(text, "\n"), nil
}

// LoadConfig will read the input reader and use the loader to source configuration files
func LoadConfig(r io.Reader, loader Loader) error {
	maps, err := readConfigs(r, 1, loader)
	if err != nil {
		return err
	}
	m := make(map[string]interface{})
	for _, config := range maps {
		for k, v := range flatten(config, "") {
			m[k] = v
		}
	}
	for k, v := range m {
		export := environmentPrefix + strings.ToUpper(k)
		env, ok := registry[export]
		if !ok {
			return fmt.Errorf("unknown key: %s (%s)", k, export)
		}
		isType, _ := env.toml()
		switch isType {
		case tomlArray:
			array, err := parseStringArray(v, true)
			if err != nil {
				return err
			}
			store.SetArray(export, array)
		case tomlInt:
			i, ok := v.(int64)
			if !ok {
				return fmt.Errorf("non-int64 found where expected: %v", v)
			}
			if i < 0 {
				return fmt.Errorf("%d is negative (not allowed here)", i)
			}
			store.SetInt64(export, i)
		case tomlBool:
			switch t := v.(type) {
			case bool:
				store.SetBool(export, t)
			default:
				return fmt.Errorf("non-bool found where expected: %v", v)
			}
		case tomlString:
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("non-string found where expected: %v", v)
			}
			store.SetString(export, os.Expand(s, os.Getenv))
		default:
			return fmt.Errorf("unknown field, can't determine type: %s (%v)", k, v)
		}

	}
	return nil
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
		including, err := parseStringArray(includes, false)
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

func parseStringArray(value interface{}, expand bool) ([]string, error) {
	var res []string
	switch t := value.(type) {
	case []interface{}:
		for _, item := range t {
			switch s := item.(type) {
			case string:
				val := s
				if expand {
					val = os.Expand(s, os.Getenv)
				}
				res = append(res, val)
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
	return LoadConfig(reader, configLoader)
}
