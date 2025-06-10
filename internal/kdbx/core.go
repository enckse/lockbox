// Package kdbx handles kdbx interactions
package kdbx

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/platform"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

var (
	errPath = errors.New("input paths must contain at LEAST 2 components")
	// AllowedFields are the same of allowed names for storing in a kdbx entry
	AllowedFields = []string{NotesField, OTPField, "Password", "URL"}
)

const (
	titleKey   = "Title"
	pathSep    = "/"
	isGlob     = "*"
	modTimeKey = "ModTime"
	// OTPField is the totp storage attribute
	OTPField = "otp"
	// NotesField is the multiline notes key
	NotesField = "Notes"
)

type (
	// EntityValues are what is stored, from an entity, into kdbx backing store
	EntityValues map[string]string
	// QuerySeq2 wraps the iteration for query entities
	QuerySeq2 iter.Seq2[Entity, error]
	// Transaction handles the overall operation of the transaction
	Transaction struct {
		file     string
		valid    bool
		exists   bool
		write    bool
		readonly bool
	}
	// Context handles operating on the underlying database
	Context struct {
		db *gokeepasslib.Database
	}
	// Entity are database objects from results and transactional changes
	Entity struct {
		Values EntityValues
		Path   string
	}
)

// Load will load a kdbx file for transactions
func Load(file string) (*Transaction, error) {
	return loadFile(file, true)
}

func loadFile(file string, must bool) (*Transaction, error) {
	if strings.TrimSpace(file) == "" {
		return nil, errors.New("no store set")
	}
	if !strings.HasSuffix(file, ".kdbx") {
		return nil, errors.New("should use a .kdbx extension")
	}
	exists := platform.PathExists(file)
	if must {
		if !exists {
			return nil, errors.New("invalid file, does not exist")
		}
	}
	ro := config.EnvReadOnly.Get()
	return &Transaction{valid: true, file: file, exists: exists, readonly: ro}, nil
}

// NewTransaction will use the underlying environment data store location
func NewTransaction() (*Transaction, error) {
	return loadFile(config.EnvStore.Get(), false)
}

func splitComponents(path string) ([]string, string, error) {
	if len(strings.Split(path, pathSep)) < 2 {
		return nil, "", errPath
	}
	if strings.HasPrefix(path, pathSep) {
		return nil, "", errors.New("path can NOT be rooted")
	}
	if strings.HasSuffix(path, pathSep) {
		return nil, "", errors.New("path can NOT end with separator")
	}
	if strings.Contains(path, pathSep+pathSep) {
		return nil, "", errors.New("unwilling to operate on path with empty segment")
	}
	title := Base(path)
	parts := strings.Split(Directory(path), pathSep)
	return parts, title, nil
}

func getCredentials(key, keyFile string) (*gokeepasslib.DBCredentials, error) {
	hasKey := len(key) > 0
	hasKeyFile := len(keyFile) > 0
	if !hasKey && !hasKeyFile {
		return nil, errors.New("key and/or keyfile must be set")
	}
	if hasKeyFile {
		if !platform.PathExists(keyFile) {
			return nil, errors.New("no keyfile found on disk")
		}
		if !hasKey {
			return gokeepasslib.NewKeyCredentials(keyFile)
		}
		return gokeepasslib.NewPasswordAndKeyCredentials(key, keyFile)
	}
	return gokeepasslib.NewPasswordCredentials(key), nil
}

func create(file, key, keyFile string) error {
	root := gokeepasslib.NewGroup()
	root.Name = "root"
	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion4())
	creds, err := getCredentials(key, keyFile)
	if err != nil {
		return err
	}
	db.Credentials = creds
	db.Content.Root = &gokeepasslib.RootData{
		Groups: []gokeepasslib.Group{root},
	}
	if err := db.LockProtectedEntries(); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return encode(f, db)
}

func encode(f *os.File, db *gokeepasslib.Database) error {
	return gokeepasslib.NewEncoder(f).Encode(db)
}

func getPathName(entry gokeepasslib.Entry) string {
	return entry.GetTitle()
}

func value(key, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func protectedValue(key, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)},
	}
}

// NewSuffix creates a new user 'name' suffix
func NewSuffix(name string) string {
	return fmt.Sprintf("%s%s", pathSep, name)
}

// NewPath creates a new storage location path.
func NewPath(segments ...string) string {
	return strings.Join(segments, pathSep)
}

// Value will read an entity value
func (e Entity) Value(key string) (string, bool) {
	if e.Values == nil {
		return "", false
	}
	val, ok := e.Values[key]
	return val, ok
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

// IsLeafAttribute indicates if a path is leaved with a certain name
func IsLeafAttribute(path, attr string) bool {
	return strings.HasSuffix(path, pathSep+attr)
}

// Collect will create a slice from an iterable set of query sequence results
func (s QuerySeq2) Collect() ([]Entity, error) {
	var entities []Entity
	for entity, err := range s {
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
