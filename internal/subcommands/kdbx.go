package subcommands

import (
	"errors"
	"flag"
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
	root := gokeepasslib.NewGroup()
	root.Name = "root"
	for _, entry := range entries {
		e := gokeepasslib.NewEntry()
		path := entry.Path
		val := entry.Value
		e.Values = append(e.Values, value("Title", filepath.Dir(path)))
		e.Values = append(e.Values, value("UserName", filepath.Base(path)))
		multi := len(strings.Split(strings.TrimSpace(val), "\n")) > 1
		if multi {
			e.Values = append(e.Values, value("Notes", val))
		} else {
			e.Values = append(e.Values, protectedValue("Password", val))
		}
		root.Entries = append(root.Entries, e)
	}
	db := &gokeepasslib.Database{
		Header:      gokeepasslib.NewHeader(),
		Credentials: gokeepasslib.NewPasswordCredentials(key),
		Content: &gokeepasslib.DBContent{
			Meta: gokeepasslib.NewMetaData(),
			Root: &gokeepasslib.RootData{
				Groups: []gokeepasslib.Group{root},
			},
		},
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
	return nil
}
