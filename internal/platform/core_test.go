package platform_test

import (
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/platform"
)

func TestPlatformList(t *testing.T) {
	if len(platform.Systems.List()) != 4 {
		t.Errorf("invalid list result")
	}
}

func TestNewPlatform(t *testing.T) {
	for _, item := range platform.Systems.List() {
		s, err := platform.NewSystem(item)
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if s != platform.System(item) {
			t.Error("mismatch on input and resulting detection")
		}
	}
}

func TestNewPlatformUnknown(t *testing.T) {
	_, err := platform.NewSystem("ifjoajei")
	if err == nil || err.Error() != "unknown platform mode" {
		t.Errorf("error expected for platform: %v", err)
	}
}
