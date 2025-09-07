package completions_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/enckse/lockbox/internal/app/completions"
	"github.com/enckse/lockbox/internal/config/store"
)

var tests = map[string]string{
	"zsh":  "typeset -A opt_args",
	"bash": "local cur opts",
}

func TestCompletions(t *testing.T) {
	for k, v := range tests {
		testCompletion(t, k, v)
	}
}

func TestCompletionReadOnly(t *testing.T) {
	defer store.Clear()
	for _, b := range []bool{true, false} {
		store.SetBool("LOCKBOX_READONLY", b)
		for k := range tests {
			v, _ := completions.Generate(k, "lb")
			res := strings.Join(v, "\n")
			for _, needs := range []string{` rm`, ` insert`, ` mv`, ` unset`} {
				has := strings.Contains(res, needs)
				if has {
					if !b {
						continue
					}
					t.Errorf("%s found, unwanted (shell %s)", needs, k)
				} else {
					if !b {
						t.Errorf("%s required, not found (shell %s)", needs, k)
					}
				}
			}
		}
	}
}

func TestFeatures(t *testing.T) {
	defer store.Clear()
	type counts struct {
		cmd     string
		with    int
		without int
	}
	for k := range tests {
		for _, feature := range []counts{{"clip", 5, 1}, {"totp", 4, 2}} {
			store.Clear()
			key := fmt.Sprintf("LOCKBOX_FEATURE_%s", strings.ToUpper(feature.cmd))
			store.SetBool(key, true)
			testCompletionFeature(t, k, feature.cmd, feature.with)
			store.SetBool(key, false)
			testCompletionFeature(t, k, feature.cmd, feature.without)
		}
	}
}

func testCompletionFeature(t *testing.T, completionMode, cmd string, expect int) {
	e := expect
	if completionMode == "bash" && cmd == "totp" {
		e++
	}
	v, _ := completions.Generate(completionMode, "lb")
	if cnt := strings.Count(strings.Join(v, "\n"), cmd); cnt != e {
		t.Errorf("completion mismatch %s: %d != %d (%s)", completionMode, cnt, expect, cmd)
	}
}

func testCompletion(t *testing.T, completionMode, need string) {
	v, err := completions.Generate(completionMode, "lb")
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if len(v) != 1 {
		t.Errorf("invalid result: %v", v)
	}
	if !strings.Contains(v[0], need) {
		t.Errorf("invalid output, bad shell generation: %v", v)
	}
}
