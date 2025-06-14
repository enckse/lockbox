package completions_test

import (
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~enckse/lockbox/internal/app/completions"
	"git.sr.ht/~enckse/lockbox/internal/config/store"
)

func TestCompletions(t *testing.T) {
	for k, v := range map[string]string{
		"zsh":  "typeset -A opt_args",
		"bash": "local cur opts",
	} {
		testCompletion(t, k, v)
	}
}

func testCompletionFeature(t *testing.T, completionMode, cmd string, expect int) {
	e := expect
	if cmd == "totp" && completionMode == "bash" {
		e++
	}
	v, _ := completions.Generate(completionMode, "lb")
	if cnt := strings.Count(strings.Join(v, "\n"), cmd); cnt != e {
		t.Errorf("completion mismatch %s: %d != %d (%s)", completionMode, cnt, expect, cmd)
	}
}

func testCompletion(t *testing.T, completionMode, need string) {
	defer store.Clear()
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
	type counts struct {
		cmd     string
		with    int
		without int
	}
	for _, feature := range []counts{{"clip", 5, 1}, {"totp", 9, 1}} {
		store.Clear()
		key := fmt.Sprintf("LOCKBOX_FEATURE_%s", strings.ToUpper(feature.cmd))
		store.SetBool(key, true)
		testCompletionFeature(t, completionMode, feature.cmd, feature.with)
		store.SetBool(key, false)
		testCompletionFeature(t, completionMode, feature.cmd, feature.without)
	}
}
