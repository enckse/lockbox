// Package kdbx handles querying a store
package kdbx

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/output"
	"github.com/tobischo/gokeepasslib/v3"
)

type (
	// QueryOptions indicates how to find entities
	QueryOptions struct {
		Criteria string
		Mode     QueryMode
		Values   ValueMode
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
	// GlobMode indicates use a glob/match to find results
	GlobMode
)

// MatchPath will try to match 1 or more elements (more elements when globbing)
func (t *Transaction) MatchPath(path string) ([]Entity, error) {
	return t.queryCollect(QueryOptions{Mode: GlobMode, Criteria: path, Values: BlankValue})
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

func forEach(offset string, groups []gokeepasslib.Group, entries []gokeepasslib.Entry, cb func(string, gokeepasslib.Entry) error) error {
	for _, g := range groups {
		o := ""
		if offset == "" {
			o = g.Name
		} else {
			o = NewPath(offset, g.Name)
		}
		if err := forEach(o, g.Groups, g.Entries, cb); err != nil {
			return err
		}
	}
	for _, e := range entries {
		if err := cb(offset, e); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transaction) queryCollect(args QueryOptions) ([]Entity, error) {
	e, err := t.QueryCallback(args)
	if err != nil {
		return nil, err
	}
	return e.Collect()
}

// Glob is the baseline query for globbing for results
func Glob(criteria, path string) (bool, error) {
	return filepath.Match(criteria, path)
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
	err := t.act(func(ctx Context) error {
		forEach("", ctx.db.Content.Root.Groups[0].Groups, ctx.db.Content.Root.Groups[0].Entries, func(offset string, entry gokeepasslib.Entry) error {
			path := getPathName(entry)
			if offset != "" {
				path = NewPath(offset, path)
			}
			if isSort {
				switch args.Mode {
				case FindMode:
					if !strings.Contains(path, args.Criteria) {
						return nil
					}
				case GlobMode:
					ok, err := Glob(args.Criteria, path)
					if err != nil {
						return err
					}
					if !ok {
						return nil
					}
				}
			} else {
				if args.Mode == ExactMode {
					if path != args.Criteria {
						return nil
					}
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
			return nil
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
	jsonHasher := func(string, string) string {
		return ""
	}
	isChecksum := false
	formatString := "%" + fmt.Sprintf("%d", (len(AllowedFields)+2)*2) + "s"
	switch jsonMode {
	case output.JSONModes.Raw:
		jsonHasher = func(val, _ string) string {
			return val
		}
	case output.JSONModes.Hash:
		isChecksum = args.Values == JSONValue
		hashLength, err := config.EnvJSONHashLength.Get()
		if err != nil {
			return nil, err
		}
		l := int(hashLength)
		jsonHasher = func(val, path string) string {
			data := fmt.Sprintf("%x", sha512.Sum512([]byte(val+path)))
			if hashLength > 0 && len(data) > l {
				data = data[0:hashLength]
			}
			return data
		}
	}
	type checksummable struct {
		value  byte
		typeof byte
	}
	return func(yield func(Entity, error) bool) {
		for _, item := range entities {
			entity := Entity{Path: item.path}
			var err error
			values := make(EntityValues)
			var checksums []checksummable
			for _, v := range item.backing.Values {
				val := ""
				raw := ""
				key := v.Key
				if args.Values != BlankValue {
					if args.Values == JSONValue {
						values["modtime"] = getValue(item.backing, modTimeKey)
					}
					val = v.Value.Content
					raw = val
					switch args.Values {
					case JSONValue:
						val = jsonHasher(val, entity.Path)
					}
				}
				if key == modTimeKey || key == titleKey {
					continue
				}
				field := strings.ToLower(key)
				if isChecksum {
					if r := jsonHasher(raw, ""); len(r) > 0 {
						checksums = append(checksums, checksummable{r[0], field[0]})
					}
				}
				values[field] = val
			}
			if isChecksum {
				var check string
				if len(checksums) > 0 {
					checksums = append(checksums, checksummable{jsonHasher(entity.Path, "")[0], byte('d')})
					slices.SortFunc(checksums, func(x, y checksummable) int {
						return int(x.typeof) - int(y.typeof)
					})
					var vals string
					for _, item := range checksums {
						vals = fmt.Sprintf("%s%s%s", vals, string(item.value), string(item.typeof))
					}
					check = strings.ReplaceAll(fmt.Sprintf(formatString, vals), " ", "0")
				}
				values[checksumKey] = check
			}
			entity.Values = values
			if !yield(entity, err) {
				return
			}
		}
	}, nil
}
