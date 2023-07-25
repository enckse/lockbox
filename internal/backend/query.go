// Package backend handles querying a store
package backend

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/tobischo/gokeepasslib/v3"
)

// MatchPath will try to match 1 or more elements (more elements when globbing)
func (t *Transaction) MatchPath(path string) ([]QueryEntity, error) {
	if !strings.HasSuffix(path, isGlob) {
		e, err := t.Get(path, BlankValue)
		if err != nil {
			return nil, err
		}
		if e == nil {
			return nil, nil
		}
		return []QueryEntity{*e}, nil
	}
	prefix := strings.TrimSuffix(path, isGlob)
	if strings.HasSuffix(prefix, pathSep) {
		return nil, errors.New("invalid match criteria, too many path separators")
	}
	return t.QueryCallback(QueryOptions{Mode: PrefixMode, Criteria: prefix + pathSep, Values: BlankValue})
}

// Get will request a singular entity
func (t *Transaction) Get(path string, mode ValueMode) (*QueryEntity, error) {
	_, _, err := splitComponents(path)
	if err != nil {
		return nil, err
	}
	e, err := t.QueryCallback(QueryOptions{Mode: ExactMode, Criteria: path, Values: mode})
	if err != nil {
		return nil, err
	}
	switch len(e) {
	case 0:
		return nil, nil
	case 1:
		return &e[0], nil
	default:
		return nil, errors.New("too many entities matched")
	}
}

func forEach(offset string, groups []gokeepasslib.Group, entries []gokeepasslib.Entry, cb func(string, gokeepasslib.Entry)) {
	for _, g := range groups {
		o := ""
		if offset == "" {
			o = g.Name
		} else {
			o = NewPath(offset, g.Name)
		}
		forEach(o, g.Groups, g.Entries, cb)
	}
	for _, e := range entries {
		cb(offset, e)
	}
}

// QueryCallback will retrieve a query based on setting
func (t *Transaction) QueryCallback(args QueryOptions) ([]QueryEntity, error) {
	if args.Mode == noneMode {
		return nil, errors.New("no query mode specified")
	}
	entities := make(map[string]QueryEntity)
	var keys []string
	isSort := args.Mode != ExactMode
	decrypt := args.Values != BlankValue
	err := t.act(func(ctx Context) error {
		forEach("", ctx.db.Content.Root.Groups[0].Groups, ctx.db.Content.Root.Groups[0].Entries, func(offset string, entry gokeepasslib.Entry) {
			path := getPathName(entry)
			if offset != "" {
				path = NewPath(offset, path)
			}
			if isSort {
				switch args.Mode {
				case FindMode:
					if !strings.Contains(path, args.Criteria) {
						return
					}
				case SuffixMode:
					if !strings.HasSuffix(path, args.Criteria) {
						return
					}
				case PrefixMode:
					if !strings.HasPrefix(path, args.Criteria) {
						return
					}
				}
			} else {
				if args.Mode == ExactMode {
					if path != args.Criteria {
						return
					}
				}
			}
			entities[path] = QueryEntity{backing: entry}
			keys = append(keys, path)
		})
		if decrypt {
			return ctx.db.UnlockProtectedEntries()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if isSort {
		sort.Strings(keys)
	}
	jsonMode := inputs.JSONDataOutputBlank
	if args.Values == JSONValue {
		m, err := inputs.ParseJSONOutput()
		if err != nil {
			return nil, err
		}
		jsonMode = m
	}
	var hashLength int
	if jsonMode == inputs.JSONDataOutputHash {
		hashLength, err = inputs.GetHashLength()
		if err != nil {
			return nil, err
		}
	}
	var results []QueryEntity
	for _, k := range keys {
		entity := QueryEntity{Path: k}
		if args.Values != BlankValue {
			e, ok := entities[k]
			if !ok {
				return nil, errors.New("failed to read entity back from map")
			}
			val := getValue(e.backing, notesKey)
			if strings.TrimSpace(val) == "" {
				val = e.backing.GetPassword()
			}
			switch args.Values {
			case JSONValue:
				data := ""
				switch jsonMode {
				case inputs.JSONDataOutputRaw:
					data = val
				case inputs.JSONDataOutputHash:
					data = fmt.Sprintf("%x", sha512.Sum512([]byte(val)))
					if hashLength > 0 && len(data) > hashLength {
						data = data[0:hashLength]
					}
				}
				t := getValue(e.backing, modTimeKey)
				s := JSON{ModTime: t, Data: data}
				m, err := json.Marshal(s)
				if err != nil {
					return nil, err
				}
				entity.Value = string(m)
			case SecretValue:
				entity.Value = val
			}
		}
		results = append(results, entity)
	}
	return results, nil
}

// NewSuffix creates a new user 'name' suffix
func NewSuffix(name string) string {
	return fmt.Sprintf("%s%s", pathSep, name)
}

// NewPath creates a new storage location path.
func NewPath(segments ...string) string {
	return strings.Join(segments, pathSep)
}

// Directory gets the offset location of the entry without the 'name'
func (e QueryEntity) Directory() string {
	return Directory(e.Path)
}

// Base will get the base name of input path
func Base(s string) string {
	parts := strings.Split(s, pathSep)
	if len(parts) == 0 {
		return s
	}
	return parts[len(parts)-1]
}

// Directory will get the directory/group for the given path
func Directory(s string) string {
	parts := strings.Split(s, pathSep)
	return NewPath(parts[0 : len(parts)-1]...)
}

func getValue(e gokeepasslib.Entry, key string) string {
	v := e.Get(key)
	if v == nil {
		return ""
	}
	return v.Value.Content
}

// IsDirectory will indicate if a path looks like a group/directory
func IsDirectory(path string) bool {
	return strings.HasSuffix(path, pathSep)
}
