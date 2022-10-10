// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"os"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/tobischo/gokeepasslib/v3"
)

// Load will load a kdbx file for transactions
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
			return nil, errors.New("invalid file, does not exist")
		}
	}
	ro, err := inputs.IsReadOnly()
	if err != nil {
		return nil, err
	}
	return &Transaction{valid: true, file: file, exists: exists, readonly: ro}, nil
}

// NewTransaction will use the underlying environment data store location
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

// pathExists indicates if a path exists.
func pathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func isTOTP(title string) (bool, error) {
	t := inputs.TOTPToken()
	if t == notesKey || t == passKey || t == titleKey {
		return false, errors.New("invalid totp field, uses restricted name")
	}
	return NewSuffix(title) == NewSuffix(t), nil
}
