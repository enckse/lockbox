package app_test

import (
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
)

func TestCompletions(t *testing.T) {
	for k, v := range map[string]string{
		"zsh":  "typeset -A opt_args",
		"fish": "set -l commands",
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

func TestProfileOptions(t *testing.T) {
	p := app.Profile{Name: "_abc-test-awera-zzz"}
	p.CanClip = true
	p.CanTOTP = true
	if len(p.Options()) != 12 {
		t.Errorf("invalid options: %v", p.Options())
	}
	p.CanClip = false
	if len(p.Options()) != 11 {
		t.Errorf("invalid options: %v", p.Options())
	}
	p.CanClip = true
	p.CanTOTP = false
	if len(p.Options()) != 11 {
		t.Errorf("invalid options: %v", p.Options())
	}
	p.CanTOTP = true
	p.ReadOnly = true
	if len(p.Options()) != 8 {
		t.Errorf("invalid options: %v", p.Options())
	}
}

func TestProfileTOTPSubOptions(t *testing.T) {
	p := app.Profile{Name: "_abc-test-awera-zzz"}
	p.CanClip = true
	if len(p.TOTPSubCommands()) != 5 {
		t.Errorf("invalid options: %v", p.TOTPSubCommands())
	}
	p.CanClip = false
	if len(p.TOTPSubCommands()) != 4 {
		t.Errorf("invalid options: %v", p.TOTPSubCommands())
	}
	p.CanClip = true
	p.ReadOnly = true
	if len(p.TOTPSubCommands()) != 4 {
		t.Errorf("invalid options: %v", p.TOTPSubCommands())
	}
}
