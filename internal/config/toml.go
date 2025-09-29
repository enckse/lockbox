package config

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/enckse/lockbox/internal/config/store"
	"github.com/enckse/lockbox/internal/output"
)

const (
	isInclude   = "include"
	maxDepth    = 10
	tomlInt     = "integer"
	tomlBool    = "boolean"
	tomlString  = "string"
	tomlArray   = "[]string"
	fileKey     = "file"
	requiredKey = "required"
)

type (
	tomlType string
	included struct {
		value    string
		required bool
	}
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
		md := item.display()
		text, err := generateDetailText(item)
		if err != nil {
			return "", err
		}
		sub = fmt.Sprintf(`%s
%s = %s
`, text, sub, md.tomlValue)
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
#
# can be an array of strings: ['file.toml']
# or an array of objects: [{%s = 'file.toml', %s = false}]
#
# this is useful to control optional includes (includes
# are required by default, set to false to allow ignoring
# files)
%s = []
`, maxDepth, fileKey, requiredKey, isInclude), "\n"} {
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
	md := data.display()
	value := md.value
	if len(value) == 0 {
		value = unset
	}
	description := strings.TrimSpace(output.TextWrap(2, env.description))
	requirement := "optional/default"
	r := strings.TrimSpace(env.requirement)
	if r != "" {
		requirement = r
	}
	var text []string
	extra := ""
	if md.canExpand {
		extra = " (shell expansions)"
	}
	for _, line := range []string{
		fmt.Sprintf("description:\n%s\n", description),
		fmt.Sprintf("requirement: %s", requirement),
		fmt.Sprintf("option: %s", strings.Join(md.allowed, "|")),
		fmt.Sprintf("default: %s", value),
		fmt.Sprintf("type: %s%s", md.tomlType, extra),
		"",
		"NOTE: the following value is NOT a default, it is an empty TOML placeholder",
	} {
		for comment := range strings.SplitSeq(line, "\n") {
			text = append(text, fmt.Sprintf("# %s", comment))
		}
	}
	return strings.Join(text, "\n"), nil
}

// Load will read the input reader and use the loader to source configuration files
func Load(r io.Reader, loader Loader) error {
	mapped, err := readConfigs(r, 1, loader)
	if err != nil {
		return err
	}
	m := make(map[string]any)
	for _, config := range mapped {
		maps.Copy(m, flatten(config, ""))
	}
	for k, v := range m {
		export := environmentPrefix + strings.ToUpper(k)
		env, ok := registry[export]
		if !ok {
			return fmt.Errorf("unknown key: %s (%s)", k, export)
		}
		md := env.display()
		switch md.tomlType {
		case tomlArray:
			array, err := parseArray(v, md.canExpand, false, func(s string, _ bool) string {
				return s
			})
			if err != nil {
				return err
			}
			store.SetArray(export, array)
		case tomlInt:
			i, ok := v.(int64)
			if !ok {
				return newTypeError("int64", v)
			}
			if i < 0 {
				return fmt.Errorf("%d is negative (not allowed here)", i)
			}
			store.SetInt64(export, i)
		case tomlBool:
			b, err := parseBool(v)
			if err != nil {
				return err
			}
			store.SetBool(export, b)
		case tomlString:
			s, ok := v.(string)
			if !ok {
				return newTypeError("string", v)
			}
			if md.canExpand {
				s = os.Expand(s, os.Getenv)
			}
			store.SetString(export, s)
		default:
			return fmt.Errorf("unknown field, can't determine type: %s (%v)", k, v)
		}

	}
	return nil
}

func newTypeError(t string, v any) error {
	return fmt.Errorf("non-%s found where %s expected: %v", t, t, v)
}

func parseBool(v any) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	default:
		return false, newTypeError("bool", v)
	}
}

func readConfigs(r io.Reader, depth int, loader Loader) ([]map[string]any, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("too many nested includes (%d > %d)", depth, maxDepth)
	}
	d := toml.NewDecoder(r)
	m := make(map[string]any)
	if _, err := d.Decode(&m); err != nil {
		return nil, err
	}
	maps := []map[string]any{m}
	includes, ok := m[isInclude]
	if ok {
		delete(m, isInclude)
		including, err := parseArray(includes, true, true, func(s string, r bool) included {
			return included{s, r}
		})
		if err != nil {
			return nil, err
		}
		if len(including) > 0 {
			for _, item := range including {
				files := []string{item.value}
				if strings.Contains(item.value, "*") {
					matched, err := filepath.Glob(item.value)
					if err != nil {
						return nil, err
					}
					files = matched
				}
				for _, file := range files {
					ok := loader.Check(file)
					if !ok {
						if item.required {
							return nil, fmt.Errorf("failed to load the included file: %s", file)
						}
						continue
					}
					reader, err := loader.Read(file)
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

func parseArray[T any](value any, expand, allowMap bool, creator func(string, bool) T) ([]T, error) {
	expander := func(s string) string {
		return s
	}
	if expand {
		expander = func(s string) string {
			return os.Expand(s, os.Getenv)
		}
	}
	var res []T
	switch t := value.(type) {
	case []any:
		for _, item := range t {
			err := fmt.Errorf("value is not valid array value: %v", item)
			val := ""
			required := true
			switch s := item.(type) {
			case string:
				val = s
			case map[string]any:
				if !allowMap {
					return nil, err
				}
				length := len(s)
				if length > 2 {
					return nil, fmt.Errorf("invalid map array, too many keys: %v", item)
				}
				file, ok := s[fileKey]
				if !ok {
					return nil, fmt.Errorf("'%s' is required, missing: %v", fileKey, item)
				}
				val, ok = file.(string)
				if !ok {
					return nil, newTypeError("string", file)
				}
				if length == 2 {
					req, ok := s[requiredKey]
					if !ok {
						return nil, fmt.Errorf("only '%s' key is allowed here: %v", requiredKey, item)
					}
					b, err := parseBool(req)
					if err != nil {
						return nil, err
					}
					required = b
				}
			default:
				return nil, err
			}
			res = append(res, creator(expander(val), required))
		}
	default:
		return nil, fmt.Errorf("value is not of array type: %v", value)
	}
	return res, nil
}

func flatten(m map[string]any, prefix string) map[string]any {
	flattened := make(map[string]any)
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "_" + k
		}

		switch to := v.(type) {
		case map[string]any:
			maps.Copy(flattened, flatten(to, key))
		default:
			flattened[key] = v
		}
	}

	return flattened
}
