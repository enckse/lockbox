package platform_test

import (
	"os"
	"testing"

	"github.com/enckse/lockbox/internal/platform"
)

func TestNewPlatform(t *testing.T) {
	for _, item := range []platform.System{platform.MacOS, platform.LinuxWayland, platform.LinuxX, platform.WindowsLinux} {
		os.Setenv("LOCKBOX_PLATFORM", string(item))
		s, err := platform.NewPlatform()
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
	_, err := platform.NewPlatform()
	if err == nil || err.Error() != "unknown platform mode" {
		t.Errorf("error expected for platform: %v", err)
	}
}
