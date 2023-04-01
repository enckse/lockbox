// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"os"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/pgl/os/paths"
	"github.com/tobischo/gokeepasslib/v3"
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
	exists := paths.Exist(file)
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

func getCredentials(key, keyFile string) (*gokeepasslib.DBCredentials, error) {
	if len(keyFile) > 0 {
		if !paths.Exist(keyFile) {
			return nil, errors.New("no keyfile found on disk")
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

func isTOTP(title string) (bool, error) {
	t := inputs.TOTPToken()
	if t == notesKey || t == passKey || t == titleKey {
		return false, errors.New("invalid totp field, uses restricted name")
	}
	return NewSuffix(title) == NewSuffix(t), nil
}
