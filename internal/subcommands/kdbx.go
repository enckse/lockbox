package subcommands

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/store"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

func value(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func protectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)},
	}
}

// ToKeepass converts the lb store to a kdbx file.
func ToKeepass(args []string) error {
	flags := flag.NewFlagSet("kdbx", flag.ExitOnError)
	file := flags.String("file", "", "file to write to")
	pass := flags.String("password", "", "password to use for the kdbx output (default is lb store key)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	fileName := *file
	if fileName == "" {
		return errors.New("no file given")
	}
	key := *pass
	if strings.TrimSpace(key) == "" {
		v, err := inputs.GetKey("", "")
		if err != nil {
			return err
		}
		key = string(v)
	}
	entries, err := DisplayCallback(DisplayOptions{All: true, Dump: true, Show: true, Store: store.NewFileSystemStore()})
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return errors.New("nothing to convert")
	}
	root := gokeepasslib.NewGroup()
	root.Name = "root"
	count := 0
	for _, entry := range entries {
		e := gokeepasslib.NewEntry()
		path := entry.Path
		val := entry.Value
		e.Values = append(e.Values, value("Title", filepath.Dir(path)))
		e.Values = append(e.Values, value("UserName", filepath.Base(path)))
		field := "Password"
		if len(strings.Split(strings.TrimSpace(val), "\n")) > 1 {
			field = "Notes"
		}
		e.Values = append(e.Values, protectedValue(field, val))
		root.Entries = append(root.Entries, e)
		count++
	}
	db := gokeepasslib.NewDatabase(gokeepasslib.WithDatabaseKDBXVersion4())
	db.Credentials = gokeepasslib.NewPasswordCredentials(key)
	db.Content.Root =
		&gokeepasslib.RootData{
			Groups: []gokeepasslib.Group{root},
		}
	if err := db.LockProtectedEntries(); err != nil {
		return err
	}
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := gokeepasslib.NewEncoder(f)
	if err := encoder.Encode(db); err != nil {
		return err
	}
	fmt.Printf("exported %d entries to %s\n", count, fileName)
	return nil
}
