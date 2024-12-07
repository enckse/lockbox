package platform_test

import (
	"testing"

	"github.com/seanenck/lockbox/internal/core"
	"github.com/seanenck/lockbox/internal/platform"
)

func TestNewPlatform(t *testing.T) {
	for _, item := range core.Platforms.List() {
		t.Setenv("LOCKBOX_PLATFORM", item)
		s, err := platform.NewPlatform()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if s != core.SystemPlatform(item) {
			t.Error("mismatch on input and resulting detection")
		}
	}
}

func TestNewPlatformUnknown(t *testing.T) {
	t.Setenv("LOCKBOX_PLATFORM", "afleaj")
	_, err := platform.NewPlatform()
	if err == nil || err.Error() != "unknown platform mode" {
		t.Errorf("error expected for platform: %v", err)
	}
}
