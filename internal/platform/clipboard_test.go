package platform_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/platform"
)

func TestNoClipboard(t *testing.T) {
	os.Setenv("LOCKBOX_CLIPMAX", "")
	os.Setenv("LOCKBOX_NOCLIP", "yes")
	_, err := platform.NewClipboard()
	if err == nil || err.Error() != "clipboard is off" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	os.Setenv("LOCKBOX_NOCLIP", "no")
	os.Setenv("LOCKBOX_PLATFORM", string(platform.LinuxWayland))
	os.Setenv("LOCKBOX_CLIPMAX", "")
	c, err := platform.NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 45 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIPMAX", "1")
	c, err = platform.NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	os.Setenv("LOCKBOX_CLIPMAX", "-1")
	_, err = platform.NewClipboard()
	if err == nil || err.Error() != "clipboard max time must be greater than 0" {
		t.Errorf("invalid max time error: %v", err)
	}
	os.Setenv("LOCKBOX_CLIPMAX", "$&(+")
	_, err = platform.NewClipboard()
	if err == nil || err.Error() != "strconv.Atoi: parsing \"$&(+\": invalid syntax" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	os.Setenv("LOCKBOX_NOCLIP", "no")
	os.Setenv("LOCKBOX_CLIPMAX", "")
	for _, item := range []platform.System{platform.MacOS, platform.LinuxWayland, platform.LinuxX, platform.WindowsLinux} {
		os.Setenv("LOCKBOX_PLATFORM", string(item))
		_, err := platform.NewClipboard()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
	}
}

func TestArgs(t *testing.T) {
	os.Setenv("LOCKBOX_PLATFORM", string(platform.WindowsLinux))
	c, _ := platform.NewClipboard()
	cmd, args := c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 {
		t.Error("invalid parse")
	}
	cmd, args = c.Args(false)
	if cmd != "powershell.exe" || len(args) != 2 || args[0] != "-command" || args[1] != "Get-Clipboard" {
		fmt.Println(args)
		t.Error("invalid parse")
	}
}
