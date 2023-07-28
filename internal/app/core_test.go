package app_test

import (
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/app"
)

func TestUsage(t *testing.T) {
	u, _ := app.Usage(false)
	if len(u) != 24 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
	u, _ = app.Usage(true)
	if len(u) != 96 {
		t.Errorf("invalid verbose usage, out of date? %d", len(u))
	}
	for _, usage := range u {
		for _, l := range strings.Split(usage, "\n") {
			if len(l) > 79 {
				t.Errorf("usage line > 79 (%d), line: %s", len(l), l)
			}
		}
	}
}
