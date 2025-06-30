package clip_test

import (
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/platform"
	"git.sr.ht/~enckse/lockbox/internal/platform/clip"
)

func TestDisabled(t *testing.T) {
	defer store.Clear()
	store.SetBool("LOCKBOX_FEATURE_CLIP", false)
	if _, err := clip.New(); err == nil || err.Error() != "clip feature is disabled" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetString("LOCKBOX_PLATFORM", string(platform.Systems.LinuxWaylandSystem))
	c, err := clip.New()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 120 {
		t.Error("invalid default")
	}
	store.SetInt64("LOCKBOX_CLIP_TIMEOUT", 1)
	c, err = clip.New()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	store.SetInt64("LOCKBOX_CLIP_TIMEOUT", -1)
	_, err = clip.New()
	if err == nil || err.Error() != "clipboard entry max time must be > 0" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	store.Clear()
	defer store.Clear()
	for _, item := range platform.Systems.List() {
		store.SetString("LOCKBOX_PLATFORM", item)
		_, err := clip.New()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
	}
}

func TestArgsOverride(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetArray("LOCKBOX_CLIP_PASTE_COMMAND", []string{"abc", "xyz", "111"})
	store.SetString("LOCKBOX_PLATFORM", string(platform.Systems.WindowsLinuxSystem))
	c, err := clip.New()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cmd, args, err := c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || err != nil {
		t.Error("invalid parse")
	}
	cmd, args, err = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || err != nil {
		t.Error("invalid parse")
	}
	store.SetArray("LOCKBOX_CLIP_COPY_COMMAND", []string{"zzz", "lll", "123"})
	c, err = clip.New()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cmd, args, err = c.Args(true)
	if cmd != "zzz" || len(args) != 2 || args[0] != "lll" || args[1] != "123" || err != nil {
		t.Error("invalid parse")
	}
	cmd, args, err = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || err != nil {
		t.Error("invalid parse")
	}
	store.Clear()
	store.SetString("LOCKBOX_PLATFORM", string(platform.Systems.WindowsLinuxSystem))
	c, err = clip.New()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cmd, args, err = c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || err != nil {
		t.Error("invalid parse")
	}
	cmd, args, err = c.Args(false)
	if cmd != "powershell.exe" || len(args) != 2 || args[0] != "-command" || args[1] != "Get-Clipboard" || err != nil {
		t.Errorf("invalid parse %s %v", cmd, args)
	}
	c = clip.Board{}
	if _, _, err := c.Args(true); err == nil || err.Error() != "command is not set (copying? true)" {
		t.Errorf("invalid error: %v", err)
	}
}
