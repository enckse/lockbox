// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.sr.ht/~enckse/lockbox/internal/config"
	"git.sr.ht/~enckse/lockbox/internal/platform"
)

type (
	// HookMode are hook operations the user can tie to
	HookMode string
	// Hook represents a runnable user-defined hook
	Hook struct {
		path    string
		mode    ActionMode
		scripts []string
		enabled bool
	}
)

const (
	internalHookEnv = "___HOOK___CALLED___"
	// HookPre are triggers BEFORE an action is performed on an entity
	HookPre HookMode = "pre"
	// HookPost are triggers AFTER an action is performed on an entity
	HookPost HookMode = "post"
)

// NewHook will create a new hook type
func NewHook(path string, a ActionMode) (Hook, error) {
	enabled := config.EnvHooksEnabled.Get()
	if !enabled || os.Getenv(internalHookEnv) != "" {
		return Hook{enabled: false}, nil
	}
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
	env := os.Environ()
	env = append(env, fmt.Sprintf("%s=1", internalHookEnv))
	for _, s := range h.scripts {
		c := exec.Command(s, string(mode), string(h.mode), h.path)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = env
		if err := c.Run(); err != nil {
			return err
		}
	}
	return nil
}
