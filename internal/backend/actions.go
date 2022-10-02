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

func (t *Transaction) act(cb action) error {
	if !t.valid {
		return errors.New("invalid transaction")
	}
	key, err := inputs.GetKey()
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
		return errors.New("kdbx must have ONE root group")
	}
	err = cb(Context{db: db})
	if err != nil {
		return err
	}
	if t.write {
		if err := db.LockProtectedEntries(); err != nil {
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
		return encode(f, db)
	}
	return err
}

func (t *Transaction) change(cb action) error {
	return t.act(func(c Context) error {
		if err := c.db.UnlockProtectedEntries(); err != nil {
			return err
		}
		t.write = true
		return cb(c)
	})
}

func (c Context) insertEntity(offset []string, title, name string, entity gokeepasslib.Entry) bool {
	return c.alterEntities(true, offset, title, name, &entity)
}

func (c Context) alterEntities(isAdd bool, offset []string, title, name string, entity *gokeepasslib.Entry) bool {
	g, e, ok := findAndDo(isAdd, NewPath(title, name), offset, entity, c.db.Content.Root.Groups[0].Groups, c.db.Content.Root.Groups[0].Entries)
	c.db.Content.Root.Groups[0].Groups = g
	c.db.Content.Root.Groups[0].Entries = e
	return ok
}

func (c Context) removeEntity(offset []string, title, name string) bool {
	return c.alterEntities(false, offset, title, name, nil)
}

func findAndDo(isAdd bool, entityName string, offset []string, opEntity *gokeepasslib.Entry, g []gokeepasslib.Group, e []gokeepasslib.Entry) ([]gokeepasslib.Group, []gokeepasslib.Entry, bool) {
	done := false
	if len(offset) == 0 {
		if isAdd {
			e = append(e, *opEntity)
		} else {
			var entries []gokeepasslib.Entry
			for _, entry := range e {
				if getPathName(entry) == entityName {
					continue
				}
				entries = append(entries, entry)
			}
			e = entries
		}
		done = true
	} else {
		name := offset[0]
		remaining := []string{}
		if len(offset) > 1 {
			remaining = offset[1:]
		}
		if isAdd {
			match := false
			for _, group := range g {
				if group.Name == name {
					match = true
				}
			}
			if !match {
				newGroup := gokeepasslib.NewGroup()
				newGroup.Name = name
				g = append(g, newGroup)
			}
		}
		var updateGroups []gokeepasslib.Group
		for _, group := range g {
			if !done && group.Name == name {
				groups, entries, ok := findAndDo(isAdd, entityName, remaining, opEntity, group.Groups, group.Entries)
				group.Entries = entries
				group.Groups = groups
				if ok {
					done = true
				}
			}
			updateGroups = append(updateGroups, group)
		}
		g = updateGroups
		if !isAdd {
			var groups []gokeepasslib.Group
			for _, group := range g {
				if group.Name == name && len(group.Entries) == 0 && len(group.Groups) == 0 {
					continue
				}
				groups = append(groups, group)
			}
			g = groups
		}
	}
	return g, e, done
}

func splitComponents(path string) ([]string, string, string, error) {
	name := filepath.Base(path)
	dir := filepath.Dir(path)
	parts := strings.Split(dir, string(os.PathSeparator))
	if len(parts) < 2 {
		return nil, "", "", errors.New("invalid component path")
	}
	return parts[:len(parts)-1], parts[len(parts)-1], name, nil
}

// Insert handles inserting a new element
func (t *Transaction) Insert(path, val string, multi bool) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("empty path not allowed")
	}
	if strings.TrimSpace(val) == "" {
		return errors.New("empty secret not allowed")
	}
	offset, title, name, err := splitComponents(path)
	if err != nil {
		return err
	}
	return t.change(func(c Context) error {
		c.removeEntity(offset, title, name)
		e := gokeepasslib.NewEntry()
		e.Values = append(e.Values, value(titleKey, title))
		e.Values = append(e.Values, value(userNameKey, name))
		field := passKey
		if multi {
			field = notesKey
		}

		e.Values = append(e.Values, protectedValue(field, val))
		c.insertEntity(offset, title, name, e)
		return nil
	})
}

// Remove handles remove an element
func (t *Transaction) Remove(entity *QueryEntity) error {
	if entity == nil {
		return errors.New("entity is empty/invalid")
	}
	offset, title, name, err := splitComponents(entity.Path)
	if err != nil {
		return err
	}
	return t.change(func(c Context) error {
		if ok := c.removeEntity(offset, title, name); !ok {
			return errors.New("failed to remove entity")
		}
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
