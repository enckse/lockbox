package app_test

import (
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestUsage(t *testing.T) {
	u, _ := app.Usage(false, "lb")
	if len(u) != 25 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
	u, _ = app.Usage(true, "lb")
	if len(u) != 108 {
		t.Errorf("invalid verbose usage, out of date? %d", len(u))
	}
}
