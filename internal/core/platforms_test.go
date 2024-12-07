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
