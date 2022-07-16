package hooks

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"github.com/enckse/lockbox/internal/misc"
)

type (
	// HookAction are specific steps that may call a hook.
	Action string
	// HookStep is the step, during command execution, when the hook was called.
	Step string
)

const (
	// RemoveHook is called when a store entry is removed.
	Remove Action = "remove"
	// InsertHook is called when a store entry is inserted.
	Insert Action = "insert"
	// PostHookStep is a hook running at the end of a command.
	PostStep Step = "post"
)

// Run executes any configured hooks.
func Run(action Action, step Step) error {
	hookDir := os.Getenv("LOCKBOX_HOOKDIR")
	if !misc.PathExists(hookDir) {
		return nil
	}
	dirs, err := os.ReadDir(hookDir)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		if !d.IsDir() {
			name := d.Name()
			cmd := exec.Command(filepath.Join(hookDir, name), string(action), string(step))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			continue
		}
		return errors.New("hook is not a file")
	}
	return nil
}
