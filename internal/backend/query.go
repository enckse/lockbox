// Package backend handles querying a store
package backend

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/output"
	"github.com/tobischo/gokeepasslib/v3"
)

type (
	// QueryOptions indicates how to find entities
	QueryOptions struct {
		PathFilter string
		Criteria   string
		Mode       QueryMode
		Values     ValueMode
	}
	// QueryMode indicates HOW an entity will be found
	QueryMode int
	// ValueMode indicates what to do with the store value of the entity
	ValueMode int
)

const (
	// BlankValue will not decrypt secrets, empty value
	BlankValue ValueMode = iota
	// SecretValue will have the raw secret onboard
	SecretValue
	// JSONValue will show entries as a JSON payload
	JSONValue
)

const (
	noneMode QueryMode = iota
	// ListMode indicates ALL entities will be listed
	ListMode
	// FindMode indicates a _contains_ search for an entity
	FindMode
	// ExactMode means an entity must MATCH the string exactly
	ExactMode
	// PrefixMode allows for entities starting with a specific value
	PrefixMode
)

// MatchPath will try to match 1 or more elements (more elements when globbing)
func (t *Transaction) MatchPath(path string) ([]Entity, error) {
	if !strings.HasSuffix(path, isGlob) {
		e, err := t.Get(path, BlankValue)
		if err != nil {
			return nil, err
		}
		if e == nil {
			return nil, nil
		}
		return []Entity{*e}, nil
	}
	prefix := strings.TrimSuffix(path, isGlob)
	if strings.HasSuffix(prefix, pathSep) {
		return nil, errors.New("invalid match criteria, too many path separators")
	}
	return t.queryCollect(QueryOptions{Mode: PrefixMode, Criteria: prefix + pathSep, Values: BlankValue})
}

// Get will request a singular entity
func (t *Transaction) Get(path string, mode ValueMode) (*Entity, error) {
	_, _, err := splitComponents(path)
	if err != nil {
		return nil, err
	}
	e, err := t.queryCollect(QueryOptions{Mode: ExactMode, Criteria: path, Values: mode})
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

func (t *Transaction) queryCollect(args QueryOptions) ([]Entity, error) {
	e, err := t.QueryCallback(args)
	if err != nil {
		return nil, err
	}
	return e.Collect()
}

// QueryCallback will retrieve a query based on setting
func (t *Transaction) QueryCallback(args QueryOptions) (QuerySeq2, error) {
	if args.Mode == noneMode {
		return nil, errors.New("no query mode specified")
	}
	type entity struct {
		path    string
		backing gokeepasslib.Entry
	}
	var entities []entity
	isSort := args.Mode != ExactMode
	decrypt := args.Values != BlankValue
	hasPathFilter := args.PathFilter != ""
	var pathFilter *regexp.Regexp
	if hasPathFilter {
		var err error
		pathFilter, err = regexp.Compile(args.PathFilter)
		if err != nil {
			return nil, err
		}
	}
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
			if hasPathFilter {
				if !pathFilter.MatchString(path) {
					return
				}
			}
			obj := entity{backing: entry, path: path}
			if isSort && len(entities) > 0 {
				i, _ := slices.BinarySearchFunc(entities, obj, func(i, j entity) int {
					return strings.Compare(i.path, j.path)
				})
				entities = slices.Insert(entities, i, obj)
			} else {
				entities = append(entities, obj)
			}
		})
		if decrypt {
			return ctx.db.UnlockProtectedEntries()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	jsonMode := output.JSONModes.Blank
	if args.Values == JSONValue {
		m, err := output.ParseJSONMode(config.EnvJSONMode.Get())
		if err != nil {
			return nil, err
		}
		jsonMode = m
	}
	jsonHasher := func(string) string {
		return ""
	}
	switch jsonMode {
	case output.JSONModes.Raw:
		jsonHasher = func(val string) string {
			return val
		}
	case output.JSONModes.Hash:
		hashLength, err := config.EnvJSONHashLength.Get()
		if err != nil {
			return nil, err
		}
		l := int(hashLength)
		jsonHasher = func(val string) string {
			data := fmt.Sprintf("%x", sha512.Sum512([]byte(val)))
			if hashLength > 0 && len(data) > l {
				data = data[0:hashLength]
			}
			return data
		}
	}
	return func(yield func(Entity, error) bool) {
		for _, item := range entities {
			entity := Entity{Path: item.path}
			var err error
			values := make(EntityValues)
			for _, v := range item.backing.Values {
				val := ""
				key := v.Key
				if args.Values != BlankValue {
					if args.Values == JSONValue {
						values["modtime"] = getValue(item.backing, modTimeKey)
					}
					val = v.Value.Content
					switch args.Values {
					case JSONValue:
						val = jsonHasher(val)
					}
				}
				if key == modTimeKey || key == titleKey {
					continue
				}
				values[strings.ToLower(key)] = val
			}
			entity.Values = values
			if !yield(entity, err) {
				return
			}
		}
	}, nil
}
