package app_test

import (
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
)

func TestUsage(t *testing.T) {
	u, _ := app.Usage(false, "lb")
	if len(u) != 25 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
	u, _ = app.Usage(true, "lb")
	if len(u) != 109 {
		t.Errorf("invalid verbose usage, out of date? %d", len(u))
	}
	for _, usage := range u {
		for _, l := range strings.Split(usage, "\n") {
			if len(l) > 80 {
				t.Errorf("usage line > 80 (%d), line: %s", len(l), l)
			}
		}
	}
}
