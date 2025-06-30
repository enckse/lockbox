package platform_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/config/store"
	"git.sr.ht/~enckse/lockbox/internal/platform"
)

type mock struct {
	err    error
	file   string
	data   string
	cmd    string
	args   []string
	pid    int
	pasted int
}

func (d *mock) WriteFile(file, data string) {
	d.file = file
	d.data = data
}

func (d *mock) ReadFile(file string) ([]byte, error) {
	if file == "falsepid" {
		return []byte("1"), nil
	}
	d.file = file
	return []byte(d.data), d.err
}

func (d *mock) Output(cmd string, args ...string) ([]byte, error) {
	d.pasted++
	d.cmd = cmd
	d.args = args
	val := fmt.Sprintf("%d", min(d.pasted, 100))
	return []byte(val), d.err
}

func (d *mock) Start(cmd string, args ...string) error {
	d.cmd = cmd
	d.args = args
	return d.err
}

func (d *mock) Getpid() int {
	return d.pid
}

func (d *mock) Copy(_ platform.Clipboard, val string) {
	d.err = fmt.Errorf("copied%s: %d", val, d.pasted)
}

func (d *mock) Sleep() {
}

func (d *mock) Loader() platform.ClipboardLoader {
	return mockLoader{name: "linux", runtime: "linux"}
}

func TestErrors(t *testing.T) {
	store.Clear()
	defer store.Clear()
	t.Setenv("WAYLAND_DISPLAY", "1")
	if err := platform.ClipboardManager(false, nil); err == nil || err.Error() != "manager is nil" {
		t.Errorf("invalid error: %v", err)
	}
	if err := platform.ClipboardManager(false, &mock{}); err == nil || err.Error() != "pidfile is unset" {
		t.Errorf("invalid error: %v", err)
	}
	store.SetString("LOCKBOX_CLIP_PIDFILE", "a")
	m := &mock{}
	m.err = errors.New("xyz")
	if err := platform.ClipboardManager(true, m); err == nil || strings.Count(err.Error(), "xyz") != 6 {
		t.Errorf("invalid error: %v", err)
	}
}

func TestStart(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetString("LOCKBOX_CLIP_PIDFILE", "a")
	t.Setenv("WAYLAND_DISPLAY", "1")
	m := &mock{}
	if err := platform.ClipboardManager(false, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.cmd != "lb" || fmt.Sprintf("%v", m.args) != "[clipmgrd]" {
		t.Errorf("invalid calls: %s %v", m.cmd, m.args)
	}
}

func TestPIDMismatch(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetString("LOCKBOX_CLIP_PIDFILE", "falsepid")
	t.Setenv("WAYLAND_DISPLAY", "1")
	m := &mock{}
	if err := platform.ClipboardManager(true, m); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestChange(t *testing.T) {
	store.Clear()
	defer store.Clear()
	store.SetString("LOCKBOX_CLIP_PIDFILE", "a")
	t.Setenv("WAYLAND_DISPLAY", "1")
	m := &mock{}
	// NOTE: 100 (count before static) + 120 (default timeout) + 1 (caused break of loop)
	if err := platform.ClipboardManager(true, m); err == nil || strings.Count(err.Error(), "copied: 221") != 6 {
		t.Errorf("invalid error: %v", err)
	}
}
