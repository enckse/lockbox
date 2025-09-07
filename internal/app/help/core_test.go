package help_test

import (
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/app/help"
)

func TestUsage(t *testing.T) {
	u, _ := help.Usage(false, "lb")
	if len(u) != 29 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
	u, _ = help.Usage(true, "lb")
	if len(u) != 130 {
		t.Errorf("invalid verbose usage, out of date? %d", len(u))
	}
	for _, usage := range u {
		for l := range strings.SplitSeq(usage, "\n") {
			if len(l) > 80 {
				t.Errorf("usage line > 80 (%d), line: %s", len(l), l)
			}
		}
	}
}
