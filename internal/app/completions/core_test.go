package completions_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app/completions"
	"github.com/seanenck/lockbox/internal/util"
)

func TestCompletions(t *testing.T) {
	for k, v := range map[string]string{
		"zsh":  "typeset -A opt_args",
		"fish": "set -f commands",
		"bash": "local cur opts",
	} {
		testCompletion(t, k, v)
	}
}

func TestConditionals(t *testing.T) {
	c := completions.NewConditionals()
	sort.Strings(c.Exported)
	need := []string{"LOCKBOX_CLIP_ENABLED", "LOCKBOX_CREDENTIALS_PASSWORD_MODE", "LOCKBOX_PWGEN_ENABLED", "LOCKBOX_READONLY", "LOCKBOX_TOTP_ENABLED"}
	if len(c.Exported) != len(need) || fmt.Sprintf("%v", c.Exported) != fmt.Sprintf("%v", need) {
		t.Errorf("invalid exports: %v", c.Exported)
	}
	fields := util.ListFields(c.Not)
	if len(fields) != len(need)+1 {
		t.Errorf("invalid fields: %v", fields)
	}
	for _, n := range need {
		value := "false"
		switch n {
		case "LOCKBOX_READONLY":
			value = "true"
		case "LOCKBOX_CREDENTIALS_PASSWORD_MODE":
			value = "ask"
		}
		found := false
		for _, f := range fields {
			if fmt.Sprintf(`[ "$%s" != "%s" ]`, n, value) == f {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("needed conditional %s not found: %v", n, fields)
		}
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
