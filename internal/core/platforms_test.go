package core_test

import (
	"testing"

	"github.com/seanenck/lockbox/internal/core"
)

func TestPlatformList(t *testing.T) {
	if len(core.Platforms.List()) != 4 {
		t.Errorf("invalid list result")
	}
}

func TestNewPlatform(t *testing.T) {
	for _, item := range core.Platforms.List() {
		s, err := core.NewPlatform(item)
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
		if s != core.SystemPlatform(item) {
			t.Error("mismatch on input and resulting detection")
		}
	}
}

func TestNewPlatformUnknown(t *testing.T) {
	_, err := core.NewPlatform("ifjoajei")
	if err == nil || err.Error() != "unknown platform mode" {
		t.Errorf("error expected for platform: %v", err)
	}
}
