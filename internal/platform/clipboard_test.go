package platform_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/platform"
)

func TestNoClipboard(t *testing.T) {
	os.Setenv("LOCKBOX_CLIP_MAX", "")
	os.Setenv("LOCKBOX_NOCLIP", "yes")
	_, err := platform.NewClipboard()
	if err == nil || err.Error() != "clipboard is off" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	os.Setenv("LOCKBOX_NOCLIP", "no")
	os.Setenv("LOCKBOX_PLATFORM", string(platform.LinuxWayland))
	os.Setenv("LOCKBOX_CLIP_MAX", "")
	c, err := platform.NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 45 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIP_MAX", "1")
	c, err = platform.NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIP_MAX", "-1")
	_, err = platform.NewClipboard()
	if err == nil || err.Error() != "clipboard max time must be greater than 0" {
		t.Errorf("invalid max time error: %v", err)
	}
	os.Setenv("LOCKBOX_CLIP_MAX", "$&(+")
	_, err = platform.NewClipboard()
	if err == nil || err.Error() != "strconv.Atoi: parsing \"$&(+\": invalid syntax" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	os.Setenv("LOCKBOX_NOCLIP", "no")
	os.Setenv("LOCKBOX_CLIP_MAX", "")
	for _, item := range []platform.System{platform.MacOS, platform.LinuxWayland, platform.LinuxX, platform.WindowsLinux} {
		os.Setenv("LOCKBOX_PLATFORM", string(item))
		_, err := platform.NewClipboard()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
	}
}

func TestArgsOverride(t *testing.T) {
	os.Setenv("LOCKBOX_CLIP_PASTE", "abc xyz 111")
	os.Setenv("LOCKBOX_PLATFORM", string(platform.WindowsLinux))
	c, _ := platform.NewClipboard()
	cmd, args := c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 {
		t.Error("invalid parse")
	}
	cmd, args = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" {
		t.Error("invalid parse")
	}
	os.Setenv("LOCKBOX_CLIP_COPY", "zzz lll 123")
	c, _ = platform.NewClipboard()
	cmd, args = c.Args(true)
	if cmd != "zzz" || len(args) != 2 || args[0] != "lll" || args[1] != "123" {
		t.Error("invalid parse")
	}
	cmd, args = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" {
		t.Error("invalid parse")
	}
	os.Setenv("LOCKBOX_CLIP_PASTE", "")
	os.Setenv("LOCKBOX_CLIP_COPY", "")
	c, _ = platform.NewClipboard()
	cmd, args = c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 {
		t.Error("invalid parse")
	}
	cmd, args = c.Args(false)
	if cmd != "powershell.exe" || len(args) != 2 || args[0] != "-command" || args[1] != "Get-Clipboard" {
		fmt.Println(args)
		t.Error("invalid parse")
	}
}
