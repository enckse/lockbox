// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

// NewHook will create a new hook type
func NewHook(path string, a ActionMode) (Hook, error) {
	if strings.TrimSpace(path) == "" {
		return Hook{}, errors.New("empty path is not allowed for hooks")
	}
	dir := inputs.EnvOrDefault(inputs.HookDirEnv, "")
	if dir == "" {
		return Hook{enabled: false}, nil
	}
	if !pathExists(dir) {
		return Hook{}, errors.New("hook directory does NOT exist")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Hook{}, err
	}
	scripts := []string{}
	for _, e := range entries {
		if e.IsDir() {
			return Hook{}, errors.New("found subdirectory in hookdir")
		}
		scripts = append(scripts, filepath.Join(dir, e.Name()))
	}
	return Hook{path: path, mode: a, enabled: len(scripts) > 0, scripts: scripts}, nil
}

// Run will execute any scripts configured as hooks
func (h Hook) Run(mode HookMode) error {
	if !h.enabled {
		return nil
	}
	for _, s := range h.scripts {
		c := exec.Command(s, string(mode), string(h.mode), h.path)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transaction) act(cb action) error {
	if !t.valid {
		return errors.New("invalid transaction")
	}
	key, err := inputs.GetKey()
	if err != nil {
		return err
	}
	k := string(key)
	file := inputs.EnvOrDefault(inputs.KeyFileEnv, "")
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

func (c Context) insertEntity(offset []string, title string, entity gokeepasslib.Entry) bool {
	return c.alterEntities(true, offset, title, &entity)
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
	title := base(path)
	parts := strings.Split(directory(path), pathSep)
	return parts, title, nil
}

// ReKey will rekey the database to a new store output
func (t *Transaction) ReKey() error {
	return t.act(func(c Context) error {
		t.write = false
		if err := inputs.SetReKey(); err != nil {
			return err
		}
		n, err := NewTransaction()
		if err != nil {
			return err
		}
		if err := c.db.UnlockProtectedEntries(); err != nil {
			return err
		}
		err = n.act(func(nCtx Context) error {
			n.write = true
			nCtx.db.Content.Root = c.db.Content.Root
			return nCtx.db.LockProtectedEntries()
		})
		if err != nil {
			return err
		}
		return c.db.LockProtectedEntries()
	})
}

// Move will move a src object to a dst location
func (t *Transaction) Move(src QueryEntity, dst string) error {
	if strings.TrimSpace(src.Path) == "" {
		return errors.New("empty path not allowed")
	}
	if strings.TrimSpace(src.Value) == "" {
		return errors.New("empty secret not allowed")
	}
	mod := inputs.EnvOrDefault(inputs.ModTimeEnv, "")
	modTime := time.Now()
	if mod != "" {
		p, err := time.Parse(inputs.ModTimeFormat, mod)
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
	action := MoveAction
	if dst == src.Path {
		action = InsertAction
	}
	hook, err := NewHook(src.Path, action)
	if err != nil {
		return err
	}
	if err := hook.Run(HookPre); err != nil {
		return err
	}
	multi := len(strings.Split(strings.TrimSpace(src.Value), "\n")) > 1
	err = t.change(func(c Context) error {
		c.removeEntity(sOffset, sTitle)
		if action == MoveAction {
			c.removeEntity(dOffset, dTitle)
		}
		e := gokeepasslib.NewEntry()
		e.Values = append(e.Values, value(titleKey, dTitle))
		field := passKey
		if multi {
			field = notesKey
		}
		v := src.Value
		ok, err := isTOTP(dTitle)
		if err != nil {
			return err
		}
		if ok {
			if multi {
				return errors.New("totp tokens can NOT be multi-line")
			}
			otp := inputs.FormatTOTP(v)
			e.Values = append(e.Values, protectedValue("otp", otp))
		}
		e.Values = append(e.Values, protectedValue(field, v))
		e.Values = append(e.Values, value(modTimeKey, modTime.Format(time.RFC3339)))
		c.insertEntity(dOffset, dTitle, e)
		return nil
	})
	if err != nil {
		return err
	}
	return hook.Run(HookPost)
}

// Insert is a move to the same location
func (t *Transaction) Insert(path, val string) error {
	return t.Move(QueryEntity{Path: path, Value: val}, path)
}

// Remove will remove a single entity
func (t *Transaction) Remove(entity *QueryEntity) error {
	if entity == nil {
		return errors.New("entity is empty/invalid")
	}
	return t.RemoveAll([]QueryEntity{*entity})
}

// RemoveAll handles removing elements
func (t *Transaction) RemoveAll(entities []QueryEntity) error {
	if len(entities) == 0 {
		return errors.New("no entities given")
	}
	removals := []removal{}
	hasHooks := false
	for _, entity := range entities {
		offset, title, err := splitComponents(entity.Path)
		if err != nil {
			return err
		}
		hook, err := NewHook(entity.Path, RemoveAction)
		if err != nil {
			return err
		}
		if err := hook.Run(HookPre); err != nil {
			return err
		}
		if hook.enabled {
			hasHooks = true
		}
		removals = append(removals, removal{parts: offset, title: title, hook: hook})
	}
	err := t.change(func(c Context) error {
		for _, entity := range removals {
			if ok := c.removeEntity(entity.parts, entity.title); !ok {
				return errors.New("failed to remove entity")
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if hasHooks {
		for _, entity := range removals {
			if err := entity.hook.Run(HookPost); err != nil {
				return err
			}
		}
	}
	return nil
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
	return entry.GetTitle()
}

func protectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)},
	}
}
