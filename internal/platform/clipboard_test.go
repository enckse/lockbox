package platform

import (
	"os"
	"testing"
)

func TestNoClipboard(t *testing.T) {
	os.Setenv("LOCKBOX_CLIPMAX", "")
	os.Setenv("LOCKBOX_NOCLIP", "yes")
	_, err := NewClipboard()
	if err == nil || err.Error() != "clipboard is off" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	os.Setenv("LOCKBOX_NOCLIP", "no")
	os.Setenv("LOCKBOX_PLATFORM", string(LinuxWayland))
	os.Setenv("LOCKBOX_CLIPMAX", "")
	c, err := NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 45 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIPMAX", "1")
	c, err = NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIPMAX", "-1")
	_, err = NewClipboard()
	if err == nil || err.Error() != "clipboard max time must be greater than 0" {
		t.Errorf("invalid max time error: %v", err)
	}
	os.Setenv("LOCKBOX_CLIPMAX", "$&(+")
	_, err = NewClipboard()
	if err == nil || err.Error() != "strconv.Atoi: parsing \"$&(+\": invalid syntax" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	os.Setenv("LOCKBOX_NOCLIP", "no")
	os.Setenv("LOCKBOX_CLIPMAX", "")
	for _, item := range []System{MacOS, LinuxWayland, LinuxX, WindowsLinux} {
		os.Setenv("LOCKBOX_PLATFORM", string(item))
		c, err := NewClipboard()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if len(c.copying) == 0 || len(c.pasting) == 0 {
			t.Error("invalid command retrieved")
		}
	}
}

func TestArgs(t *testing.T) {
	c := Clipboard{copying: []string{"cp"}, pasting: []string{"paste", "with", "args"}}
	cmd, args := c.Args(true)
	if cmd != "cp" || len(args) != 0 {
		t.Error("invalid parse")
	}
	cmd, args = c.Args(false)
	if cmd != "paste" || len(args) != 2 || args[0] != "with" || args[1] != "args" {
		t.Error("invalid parse")
	}
}
