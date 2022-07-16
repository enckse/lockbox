package clip

import (
	"os"
	"testing"
)

func TestNoClipboard(t *testing.T) {
	os.Setenv("LOCKBOX_CLIPMAX", "")
	os.Setenv("LOCKBOX_CLIPMODE", "off")
	_, err := NewCommands()
	if err == nil || err.Error() != "clipboard is unavailable" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	os.Setenv("LOCKBOX_CLIPMODE", pbClipMode)
	os.Setenv("LOCKBOX_CLIPMAX", "")
	c, err := NewCommands()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 45 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIPMAX", "1")
	c, err = NewCommands()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIPMAX", "-1")
	_, err = NewCommands()
	if err == nil || err.Error() != "clipboard max time must be greater than 0" {
		t.Errorf("invalid max time error: %v", err)
	}
	os.Setenv("LOCKBOX_CLIPMAX", "$&(+")
	_, err = NewCommands()
	if err == nil || err.Error() != "strconv.Atoi: parsing \"$&(+\": invalid syntax" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	os.Setenv("LOCKBOX_CLIPMAX", "")
	for _, item := range []string{pbClipMode, xClipMode, waylandClipMode, wslMode} {
		os.Setenv("LOCKBOX_CLIPMODE", item)
		c, err := NewCommands()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if len(c.copying) == 0 || len(c.pasting) == 0 {
			t.Error("invalid command retrieved")
		}
	}
}

func TestArgs(t *testing.T) {
	c := Commands{copying: []string{"cp"}, pasting: []string{"paste", "with", "args"}}
	cmd, args := c.Args(true)
	if cmd != "cp" || len(args) != 0 {
		t.Error("invalid parse")
	}
	cmd, args = c.Args(false)
	if cmd != "paste" || len(args) != 2 || args[0] != "with" || args[1] != "args" {
		t.Error("invalid parse")
	}
}
