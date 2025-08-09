package platform_test

import (
	"fmt"
	"testing"

	"github.com/enckse/lockbox/internal/config/store"
	"github.com/enckse/lockbox/internal/platform"
)

type mockLoader struct {
	err     error
	name    string
	runtime string
}

func (m mockLoader) Name() (string, error) {
	return m.name, m.err
}

func (m mockLoader) Runtime() string {
	return m.runtime
}

func TestDisabled(t *testing.T) {
	defer store.Clear()
	store.SetBool("LOCKBOX_FEATURE_CLIP", false)
	if _, err := platform.NewClipboard(mockLoader{}); err == nil || err.Error() != "clip feature is disabled" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestInstance(t *testing.T) {
	store.Clear()
	defer store.Clear()
	fxn := func(runtime, name, c, e string) {
		l := mockLoader{runtime: runtime, name: name}
		b, err := platform.NewClipboard(l)
		if err != nil {
			if err.Error() != e {
				t.Errorf("invalid error: %v", err)
			}
			return
		}
		if fmt.Sprintf("%v", b) != c {
			t.Errorf("invalid copy: %s != %v", c, b)
		}
	}
	fxn("darwin", "", "[pbcopy]", "")
	fxn("linux", "microsoft", "[clip.exe]", "")
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	fxn("linux", "linux", "", "unable to detect linux clipboard")
	t.Setenv("DISPLAY", "1")
	t.Setenv("WAYLAND_DISPLAY", "1")
	fxn("linux", "linux", "[wl-copy]", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	fxn("linux", "linux", "[xclip]", "")
}

func TestCopy(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetArray("LOCKBOX_CLIP_COPY", []string{})
	if _, err := platform.NewClipboard(mockLoader{}); err == nil || err.Error() != "clipboard is unavailable" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetArray("LOCKBOX_CLIP_COPY", []string{"x"})
	c, err := platform.NewClipboard(mockLoader{name: "microsoft", runtime: "linux"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", c) != "[x]" {
		t.Errorf("invalid override: %v", c)
	}
	store.SetArray("LOCKBOX_CLIP_COPY", []string{"x", "y", "z"})
	c, err = platform.NewClipboard(mockLoader{name: "microsoft", runtime: "linux"})
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if fmt.Sprintf("%v", c) != "[x y z]" {
		t.Errorf("invalid override: %v", c)
	}
	c = platform.Clipboard{}
	if err := c.CopyTo(""); err == nil || err.Error() != "copy command is not set" {
		t.Errorf("invalid error: %v", err)
	}
	c = platform.Clipboard{"echo"}
	if err := c.CopyTo(""); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}
