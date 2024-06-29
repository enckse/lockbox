package backend_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/backend"
)

func TestHooks(t *testing.T) {
	os.Setenv("LOCKBOX_HOOKDIR", "")
	h, err := backend.NewHook("a", backend.InsertAction)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := h.Run(backend.HookPre); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := backend.NewHook("", backend.InsertAction); err.Error() != "empty path is not allowed for hooks" {
		t.Errorf("wrong error: %v", err)
	}
	os.Setenv("LOCKBOX_HOOKDIR", "is_garbage")
	if _, err := backend.NewHook("b", backend.InsertAction); err.Error() != "hook directory does NOT exist" {
		t.Errorf("wrong error: %v", err)
	}
	testPath := filepath.Join("testdata", "hooks.kdbx")
	os.RemoveAll(testPath)
	if err := os.MkdirAll(testPath, 0o755); err != nil {
		t.Errorf("failed, mkdir: %v", err)
	}
	os.Setenv("LOCKBOX_HOOKDIR", testPath)
	h, err = backend.NewHook("a", backend.InsertAction)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := h.Run(backend.HookPre); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	sub := filepath.Join(testPath, "subdir")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Errorf("failed, mkdir sub: %v", err)
	}
	if _, err := backend.NewHook("b", backend.InsertAction); err.Error() != "found subdirectory in hookdir" {
		t.Errorf("wrong error: %v", err)
	}
	if err := os.RemoveAll(sub); err != nil {
		t.Errorf("failed rmdir: %v", err)
	}
	script := filepath.Join(testPath, "testscript")
	if err := os.WriteFile(script, []byte{}, 0o644); err != nil {
		t.Errorf("unable to write script: %v", err)
	}
	h, err = backend.NewHook("a", backend.InsertAction)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if err := h.Run(backend.HookPre); strings.Contains("fork/exec", err.Error()) {
		t.Errorf("wrong error: %v", err)
	}
}
