package clip_test

import (
	"testing"

	"codeberg.org/enckse/lockbox/internal/config/store"
	"codeberg.org/enckse/lockbox/internal/platform"
	"codeberg.org/enckse/lockbox/internal/platform/clip"
)

func TestNoClipboard(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetBool("LOCKBOX_CLIP_OSC52", false)
	store.SetBool("LOCKBOX_CLIP_ENABLED", false)
	_, err := clip.New()
	if err == nil || err.Error() != "clipboard is off" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetBool("LOCKBOX_CLIP_OSC52", false)
	store.SetBool("LOCKBOX_CLIP_ENABLED", true)
	store.SetString("LOCKBOX_PLATFORM", string(platform.Systems.LinuxWaylandSystem))
	c, err := clip.New()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 45 {
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
	if err == nil || err.Error() != "clipboard max time must be > 0" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetBool("LOCKBOX_CLIP_OSC52", false)
	store.SetBool("LOCKBOX_CLIP_ENABLED", true)
	for _, item := range platform.Systems.List() {
		store.SetString("LOCKBOX_PLATFORM", item)
		_, err := clip.New()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
	}
}

func TestOSC52(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetBool("LOCKBOX_CLIP_OSC52", true)
	c, _ := clip.New()
	_, _, ok := c.Args(true)
	if ok {
		t.Error("invalid clipboard, should be an internal call")
	}
	_, _, ok = c.Args(false)
	if ok {
		t.Error("invalid clipboard, should be an internal call")
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
	cmd, args, ok := c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || !ok {
		t.Error("invalid parse")
	}
	store.SetArray("LOCKBOX_CLIP_COPY_COMMAND", []string{"zzz", "lll", "123"})
	c, err = clip.New()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cmd, args, ok = c.Args(true)
	if cmd != "zzz" || len(args) != 2 || args[0] != "lll" || args[1] != "123" || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || !ok {
		t.Error("invalid parse")
	}
	store.Clear()
	store.SetString("LOCKBOX_PLATFORM", string(platform.Systems.WindowsLinuxSystem))
	c, err = clip.New()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	cmd, args, ok = c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "powershell.exe" || len(args) != 2 || args[0] != "-command" || args[1] != "Get-Clipboard" || !ok {
		t.Errorf("invalid parse %s %v", cmd, args)
	}
}
