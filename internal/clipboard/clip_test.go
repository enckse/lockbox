package clipboard

import (
	"os"
	"testing"
)

func TestNoClipboard(t *testing.T) {
	os.Setenv("LOCKBOX_CLIPMODE", "off")
	_, err := NewCommands()
	if err == nil || err.Error() != "clipboard is unavailable" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	for _, item := range []string{pbClipMode, xClipMode, waylandClipMode, wslMode} {
		os.Setenv("LOCKBOX_CLIPMODE", item)
		c, err := NewCommands()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if len(c.Copy) == 0 || len(c.Paste) == 0 {
			t.Error("invalid command retrieved")
		}
	}
}
