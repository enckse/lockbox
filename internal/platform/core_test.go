package platform

import (
	"os"
	"testing"
)

func TestNewPlatform(t *testing.T) {
	for _, item := range []System{MacOS, LinuxWayland, LinuxX, WindowsLinux} {
		os.Setenv("LOCKBOX_PLATFORM", string(item))
		s, err := NewPlatform()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if s != item {
			t.Error("mismatch on input and resulting detection")
		}
	}
}

func TestNewPlatformUnknown(t *testing.T) {
	os.Setenv("LOCKBOX_PLATFORM", "afleaj")
	_, err := NewPlatform()
	if err == nil || err.Error() != "unknown platform mode" {
		t.Errorf("error expected for platform: %v", err)
	}
}
