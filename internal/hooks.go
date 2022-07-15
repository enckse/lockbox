package internal

import (
	"os"
	"os/exec"
	"path/filepath"
)

type (
	HookAction string
	HookStep string
)

const (
	RemoveHook HookAction = "remove"
	InsertHook HookAction = "insert"
	PostHookStep HookStep = "post"
)

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
