package internal

import (
	"os"
	"os/exec"
	"path/filepath"
)

type (
	// HookAction are specific steps that may call a hook.
	HookAction string
	// HookStep is the step, during command execution, when the hook was called.
	HookStep string
)

const (
	// RemoveHook is called when a store entry is removed.
	RemoveHook HookAction = "remove"
	// InsertHook is called when a store entry is inserted.
	InsertHook HookAction = "insert"
	// PostHookStep is a hook running at the end of a command.
	PostHookStep HookStep = "post"
)

// Hooks executes any configured hooks.
func Hooks(store string, action HookAction, step HookStep) error {
	hookDir := os.Getenv("LOCKBOX_HOOKDIR")
	if !PathExists(hookDir) {
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
		return NewLockboxError("hook is not a file")
	}
	return nil
}
