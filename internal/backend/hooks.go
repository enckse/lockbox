// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/enckse/lockbox/internal/system"
)

// NewHook will create a new hook type
func NewHook(path string, a ActionMode) (Hook, error) {
	if strings.TrimSpace(path) == "" {
		return Hook{}, errors.New("empty path is not allowed for hooks")
	}
	dir := system.EnvironOrDefault(inputs.HookDirEnv, "")
	if dir == "" {
		return Hook{enabled: false}, nil
	}
	if !system.PathExists(dir) {
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
