// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/enckse/lockbox/internal/config"
	"github.com/enckse/lockbox/internal/platform"
)

type (
	// HookMode are hook operations the user can tie to
	HookMode string
	// Hook represents a runnable user-defined hook
	Hook struct {
		path    string
		mode    ActionMode
		enabled bool
		scripts []string
	}
)

const (
	// HookPre are triggers BEFORE an action is performed on an entity
	HookPre HookMode = "pre"
	// HookPost are triggers AFTER an action is performed on an entity
	HookPost HookMode = "post"
)

// NewHook will create a new hook type
func NewHook(path string, a ActionMode) (Hook, error) {
	if strings.TrimSpace(path) == "" {
		return Hook{}, errors.New("empty path is not allowed for hooks")
	}
	dir := config.EnvHookDir.Get()
	if dir == "" {
		return Hook{enabled: false}, nil
	}
	if !platform.PathExists(dir) {
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
