package clip_test

import (
	"fmt"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/platform/clip"
)

type mockLoader struct {
	err     error
	name    string
	runtime string
	full    bool
}

func (m mockLoader) Name() (string, error) {
	return m.name, m.err
}

func (m mockLoader) Runtime() string {
	return m.runtime
}

func (m mockLoader) Complete() bool {
	return m.full
}

func TestDisabled(t *testing.T) {
	defer store.Clear()
	store.SetBool("LOCKBOX_FEATURE_CLIP", false)
	if _, err := clip.New(mockLoader{}); err == nil || err.Error() != "clip feature is disabled" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	store.Clear()
	defer store.Clear()
	t.Setenv("WAYLAND_DISPLAY", "1")
	loader := mockLoader{name: "linux", runtime: "linux"}
	c, err := clip.New(loader)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.MaxTime != 120 {
		t.Error("invalid default")
	}
	store.SetInt64("LOCKBOX_CLIP_TIMEOUT", 1)
	c, err = clip.New(loader)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	store.SetInt64("LOCKBOX_CLIP_TIMEOUT", -1)
	c, err = clip.New(loader)
	if err == nil || err.Error() != "clipboard entry max time must be > 0" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestInstance(t *testing.T) {
	store.Clear()
	defer store.Clear()
	fxn := func(runtime, name, c, p, e string) {
		l := mockLoader{runtime: runtime, name: name}
		b, err := clip.New(l)
		if err != nil {
			if err.Error() != e {
				t.Errorf("invalid error: %v", err)
			}
			return
		}
		cmd, args, _ := b.Args(true)
		copying := fmt.Sprintf("%s (%v)", cmd, args)
		cmd, args, _ = b.Args(false)
		pasting := fmt.Sprintf("%s (%v)", cmd, args)
		if copying != c {
			t.Errorf("invalid copy: %s != %s", c, copying)
		}
		if pasting != p {
			t.Errorf("invalid copy: %s != %s", p, pasting)
		}
	}
	fxn("darwin", "", "pbcopy ([])", "pbpaste ([])", "")
	fxn("linux", "microsoft", "clip.exe ([])", "powershell.exe ([-command Get-Clipboard])", "")
	fxn("linux", "linux", "", "", "unable to detect linux clipboard")
	t.Setenv("DISPLAY", "1")
	t.Setenv("WAYLAND_DISPLAY", "1")
	fxn("linux", "linux", "wl-copy ([])", "wl-paste ([])", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	fxn("linux", "linux", "xclip ([])", "xclip ([-o])", "")
}

func TestFullPartial(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetArray("LOCKBOX_CLIP_COPY_COMMAND", []string{"abc", "xyz", "111"})
	if _, err := clip.New(mockLoader{}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := clip.New(mockLoader{full: true}); err == nil || err.Error() != "clipboard is unavailable" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetArray("LOCKBOX_CLIP_PASTE_COMMAND", []string{"abc", "xyz", "111"})
	if _, err := clip.New(mockLoader{full: true}); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestArgsOverride(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetArray("LOCKBOX_CLIP_PASTE_COMMAND", []string{"abc", "xyz", "111"})
	c, err := clip.New(mockLoader{name: "microsoft", runtime: "linux"})
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
	c, err = clip.New(mockLoader{})
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
	c, err = clip.New(mockLoader{name: "microsoft", runtime: "linux"})
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
