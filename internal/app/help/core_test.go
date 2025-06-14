package help_test

import (
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app/help"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func TestUsage(t *testing.T) {
	u, _ := help.Usage(false, "lb")
	if len(u) != 27 {
		t.Errorf("invalid usage, out of date? %d", len(u))
	}
	u, _ = help.Usage(true, "lb")
	if len(u) != 117 {
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

func TestFlags(t *testing.T) {
	defer store.Clear()
	for _, feature := range []string{"clip", "totp", "color"} {
		store.Clear()
		key := fmt.Sprintf("LOCKBOX_FEATURE_%s", strings.ToUpper(feature))
		store.SetBool(key, true)
		u, _ := help.Usage(true, "lb")
		if !strings.Contains(strings.Join(u, "\n"), feature) {
			t.Errorf("verbose help lacks: %s", feature)
		}
		store.SetBool(key, false)
		u, _ = help.Usage(true, "lb")
		if strings.Contains(strings.Join(u, "\n"), feature) {
			t.Errorf("verbose help has: %s", feature)
		}
	}
}
