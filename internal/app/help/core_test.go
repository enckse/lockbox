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
	if len(u) != 121 {
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

func TestReadOnly(t *testing.T) {
	defer store.Clear()
	check := func(require bool) {
		u, _ := help.Usage(true, "lb")
		text := strings.Join(u, "\n")
		for _, need := range []string{"  mv ", "  rm ", "[rekey]", "  insert ", "  unset ", "[globs]"} {
			has := strings.Contains(text, need)
			if has {
				if require {
					continue
				}
				t.Errorf("has unwanted text: %s", need)
			} else {
				if require {
					t.Errorf("missing required text: %s", need)
				}
			}
		}
	}

	check(true)
	store.SetBool("LOCKBOX_READONLY", true)
	check(false)
}
