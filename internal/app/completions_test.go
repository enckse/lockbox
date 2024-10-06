package app_test

import (
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
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

func testCompletion(t *testing.T, completionMode, need string) {
	v, err := app.GenerateCompletions(completionMode, "lb")
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
