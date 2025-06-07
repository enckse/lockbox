// Package kdbx handles kdbx interactions
package kdbx

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/platform"
	"github.com/tobischo/gokeepasslib/v3"
)

type (
	action func(t Context) error
)

func (t *Transaction) act(cb action) error {
	if !t.valid {
		return errors.New("invalid transaction")
	}
	key, err := config.NewKey(config.DefaultKeyMode)
	if err != nil {
		return err
	}
	k, err := key.Read(platform.ReadInteractivePassword)
	if err != nil {
		return err
	}
	file := config.EnvKeyFile.Get()
	if !t.exists {
		if err := create(t.file, k, file); err != nil {
			return err
		}
	}
	f, err := os.Open(t.file)
	if err != nil {
		return err
	}
	defer f.Close()
	db := gokeepasslib.NewDatabase()
	creds, err := getCredentials(k, file)
	if err != nil {
		return err
	}
	db.Credentials = creds
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

// ReKey will change the credentials on a database
func (t *Transaction) ReKey(pass, keyFile string) error {
	creds, err := getCredentials(pass, keyFile)
	if err != nil {
		return err
	}
	return t.change(func(c Context) error {
		c.db.Credentials = creds
		return nil
	})
}

func (t *Transaction) change(cb action) error {
	if t.readonly {
		return errors.New("unable to alter database in readonly mode")
	}
	return t.act(func(c Context) error {
		if err := c.db.UnlockProtectedEntries(); err != nil {
			return err
		}
		t.write = true
		return cb(c)
	})
}

func (c Context) alterEntities(isAdd bool, offset []string, title string, entity *gokeepasslib.Entry) bool {
	g, e, ok := findAndDo(isAdd, title, offset, entity, c.db.Content.Root.Groups[0].Groups, c.db.Content.Root.Groups[0].Entries)
	c.db.Content.Root.Groups[0].Groups = g
	c.db.Content.Root.Groups[0].Entries = e
	return ok
}

func (c Context) removeEntity(offset []string, title string) bool {
	return c.alterEntities(false, offset, title, nil)
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

// Move will move a src object to a dst location
func (t *Transaction) Move(src *Entity, dst string) error {
	if src == nil {
		return errors.New("source entity is not set")
	}
	if strings.TrimSpace(src.Path) == "" {
		return errors.New("empty path not allowed")
	}
	if len(src.Values) == 0 {
		return errors.New("empty secrets not allowed")
	}
	values := make(map[string]string)
	for k, v := range src.Values {
		found := false
		for _, mapping := range AllowedFields {
			if strings.EqualFold(k, mapping) {
				values[mapping] = v
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown entity field: %s", k)
		}
	}
	mod := config.EnvDefaultModTime.Get()
	modTime := time.Now()
	if mod != "" {
		p, err := time.Parse(config.ModTimeFormat, mod)
		if err != nil {
			return err
		}
		modTime = p
	}
	dOffset, dTitle, err := splitComponents(dst)
	if err != nil {
		return err
	}
	sOffset, sTitle, err := splitComponents(src.Path)
	if err != nil {
		return err
	}
	isMove := dst != src.Path
	return t.change(func(c Context) error {
		c.removeEntity(sOffset, sTitle)
		if isMove {
			c.removeEntity(dOffset, dTitle)
		}
		e := gokeepasslib.NewEntry()
		e.Values = append(e.Values, value(titleKey, dTitle))
		e.Values = append(e.Values, value(modTimeKey, modTime.Format(time.RFC3339)))
		for k, v := range values {
			val := v
			switch k {
			case OTPField, PasswordField:
				if strings.Contains(val, "\n") {
					return fmt.Errorf("%s can NOT be multi-line", strings.ToLower(k))
				}
				if k == OTPField {
					val = config.EnvTOTPFormat.Get(v)
				}
			}
			e.Values = append(e.Values, protectedValue(k, val))
		}
		c.alterEntities(true, dOffset, dTitle, &e)
		return nil
	})
}

// Insert is a move to the same location
func (t *Transaction) Insert(path string, val EntityValues) error {
	return t.Move(&Entity{Path: path, Values: val}, path)
}

// Remove will remove a single entity
func (t *Transaction) Remove(entity *Entity) error {
	if entity == nil {
		return errors.New("entity is empty/invalid")
	}
	return t.RemoveAll([]Entity{*entity})
}

// RemoveAll handles removing elements
func (t *Transaction) RemoveAll(entities []Entity) error {
	if len(entities) == 0 {
		return errors.New("no entities given")
	}
	type removal struct {
		title string
		parts []string
	}
	removals := []removal{}
	for _, entity := range entities {
		offset, title, err := splitComponents(entity.Path)
		if err != nil {
			return err
		}
		removals = append(removals, removal{parts: offset, title: title})
	}
	return t.change(func(c Context) error {
		for _, entity := range removals {
			if ok := c.removeEntity(entity.parts, entity.title); !ok {
				return errors.New("failed to remove entity")
			}
		}
		return nil
	})
}
