// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

const (
	userNameKey = "UserName"
	notesKey    = "Notes"
	titleKey    = "Title"
	passKey     = "Password"
)

type (
	Action      func(t Context) error
	Transaction struct {
		valid  bool
		file   string
		exists bool
		db     *gokeepasslib.Database
		write  bool
	}
	Context struct {
		db *gokeepasslib.Database
	}
)

func Load(file string) (*Transaction, error) {
	return loadFile(file, true)
}

func loadFile(file string, must bool) (*Transaction, error) {
	if !strings.HasSuffix(file, ".kdbx") {
		return nil, errors.New("should use a .kdbx extension")
	}
	exists := pathExists(file)
	if must {
		if !exists {
			return nil, errors.New("invalid file, does not exists")
		}
	}
	return &Transaction{valid: true, file: file, exists: exists}, nil
}

func NewTransaction() (*Transaction, error) {
	return loadFile(os.Getenv(inputs.StoreEnv), false)
}

func create(file, key string) error {
	root := gokeepasslib.NewGroup()
	root.Name = "root"
	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion4())
	db.Credentials = gokeepasslib.NewPasswordCredentials(key)
	db.Content.Root =
		&gokeepasslib.RootData{
			Groups: []gokeepasslib.Group{root},
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

func (t *Transaction) Act(cb Action) error {
	if !t.valid {
		return errors.New("invalid transaction")
	}
	key, err := inputs.GetKey("", "")
	if err != nil {
		return err
	}
	k := string(key)
	if !t.exists {
		if err := create(t.file, k); err != nil {
			return err
		}
	}
	f, err := os.Open(t.file)
	if err != nil {
		return err
	}
	defer f.Close()
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(k)
	if err := gokeepasslib.NewDecoder(f).Decode(db); err != nil {
		return err
	}
	if len(db.Content.Root.Groups) != 1 {
		return errors.New("kdbx must only have ONE root group")
	}
	t.db = db
	cErr := cb(Context{db: t.db})
	if t.write {
		if err := t.db.LockProtectedEntries(); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		f, err = os.Create(t.file)
		if err != nil {
			return err
		}
		defer f.Close()
		return encode(f, t.db)
	}
	return cErr
}

func (t *Transaction) change(cb Action) error {
	return t.Act(func(c Context) error {
		if err := c.db.UnlockProtectedEntries(); err != nil {
			return err
		}
		t.write = true
		return cb(c)
	})
}

func (t *Transaction) Insert(path, val string, multi bool) error {
	return t.change(func(c Context) error {
		e := gokeepasslib.NewEntry()
		e.Values = append(e.Values, value(titleKey, filepath.Dir(path)))
		e.Values = append(e.Values, value(userNameKey, filepath.Base(path)))
		field := passKey
		if multi {
			field = notesKey
		}

		e.Values = append(e.Values, protectedValue(field, val))
		c.db.Content.Root.Groups[0].Entries = append(c.db.Content.Root.Groups[0].Entries, e)
		return nil
	})
}

func (t *Transaction) Remove(entity *QueryEntity) error {
	if entity == nil {
		return errors.New("entity is empty/invalid")
	}
	return t.change(func(c Context) error {
		entries := c.db.Content.Root.Groups[0].Entries
		if entity.Index >= len(entries) {
			return errors.New("index out of bounds")
		}
		e := entries[entity.Index]
		n := getPathName(e)
		if n != entity.Path {
			return errors.New("validation failed, index/name mismatch")
		}
		c.db.Content.Root.Groups[0].Entries = append(entries[:entity.Index], entries[entity.Index+1:]...)
		return nil
	})
}

func getValue(e gokeepasslib.Entry, key string) string {
	v := e.Get(key)
	if v == nil {
		return ""
	}
	return v.Value.Content
}

func value(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func getPathName(entry gokeepasslib.Entry) string {
	return filepath.Join(entry.GetTitle(), getValue(entry, userNameKey))
}

func protectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)},
	}
}

// pathExists indicates if a path exists.
func pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}